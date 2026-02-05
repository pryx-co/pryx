use super::auth::auth_middleware;
use super::handlers::{
    audit_list_handler, channel_create_handler, channel_delete_handler, channel_get_handler,
    channel_test_handler, channel_update_handler, channels_list_handler, config_handler,
    cost_summary_handler, health_handler, mcp_create_handler, mcp_delete_handler, mcp_get_handler,
    mcp_list_handler, mcp_update_handler, models_handler, policy_create_handler,
    policy_delete_handler, policy_get_handler, policy_list_handler, policy_update_handler,
    providers_handler, skills_handler,
};
use super::websocket::handle_socket;
use crate::server::ServerConfig;
use axum::{
    http::{header, StatusCode},
    middleware,
    response::IntoResponse,
    routing::{get, post},
    Router,
};
use std::path::PathBuf;
use tower_http::services::ServeDir;

async fn static_files_handler(
    uri: axum::http::Uri,
    axum::extract::State(config): axum::extract::State<super::ServerConfig>,
) -> axum::response::Response {
    // Resolve canonical base directory at startup
    let base_dir = match std::fs::canonicalize(&config.static_files_path) {
        Ok(path) => path,
        Err(e) => {
            log::error!("Failed to resolve static files base directory: {}", e);
            return (
                StatusCode::INTERNAL_SERVER_ERROR,
                "Server configuration error",
            )
                .into_response();
        }
    };

    let path = uri.path().trim_start_matches('/');

    // Sanitize path: remove leading '/' and reject ".." segments
    let sanitized_path: PathBuf = path
        .split('/')
        .filter(|segment| !segment.is_empty() && *segment != "..")
        .collect();

    // Reject paths with ".." after the filter (edge case)
    if path.contains("..") {
        return (StatusCode::BAD_REQUEST, "Invalid path").into_response();
    }

    let target_path = if sanitized_path.components().next().is_none()
        || path.is_empty()
        || path == "index.html"
    {
        base_dir.join("index.html")
    } else {
        base_dir.join(&sanitized_path)
    };

    // Verify the resolved path is within the base directory
    if !target_path.starts_with(&base_dir) {
        return (StatusCode::FORBIDDEN, "Access denied").into_response();
    }

    // Use async tokio fs operations
    match tokio::fs::metadata(&target_path).await {
        Ok(metadata) => {
            if metadata.is_file() {
                match tokio::fs::read(&target_path).await {
                    Ok(content) => {
                        let mime_type = mime_guess::from_path(&target_path).first_or_octet_stream();
                        ([("Content-Type", mime_type.as_ref())], content).into_response()
                    }
                    Err(e) => {
                        log::error!("Failed to read static file: {}", e);
                        (StatusCode::INTERNAL_SERVER_ERROR, "Failed to read file").into_response()
                    }
                }
            } else {
                (StatusCode::NOT_FOUND, "File not found").into_response()
            }
        }
        Err(e) if e.kind() == std::io::ErrorKind::NotFound => {
            (StatusCode::NOT_FOUND, "File not found").into_response()
        }
        Err(e) => {
            log::error!("Error checking static file: {}", e);
            (StatusCode::INTERNAL_SERVER_ERROR, "Server error").into_response()
        }
    }
}

pub fn app_router(config: ServerConfig) -> Router {
    let api_routes = Router::new()
        .route("/health", get(health_handler))
        .route("/skills", get(skills_handler))
        .route("/config", get(config_handler))
        .route("/providers", get(providers_handler))
        .route("/models", get(models_handler))
        // Channels
        .route(
            "/channels",
            get(channels_list_handler).post(channel_create_handler),
        )
        .route(
            "/channels/:id",
            get(channel_get_handler)
                .put(channel_update_handler)
                .delete(channel_delete_handler),
        )
        .route("/channels/:id/test", post(channel_test_handler))
        // MCP
        .route("/mcp", get(mcp_list_handler).post(mcp_create_handler))
        .route(
            "/mcp/:id",
            get(mcp_get_handler)
                .put(mcp_update_handler)
                .delete(mcp_delete_handler),
        )
        // Policies
        .route(
            "/policies",
            get(policy_list_handler).post(policy_create_handler),
        )
        .route(
            "/policies/:id",
            get(policy_get_handler)
                .put(policy_update_handler)
                .delete(policy_delete_handler),
        )
        // Audit & Cost
        .route("/audit/logs", get(audit_list_handler))
        .route("/cost/summary", get(cost_summary_handler))
        .layer(middleware::from_fn(auth_middleware))
        .with_state(config.clone());

    Router::new()
        .route("/", get(root_handler))
        .nest("/api", api_routes)
        // WS
        .route("/ws", get(ws_upgrade_handler))
        .nest_service("/static", ServeDir::new(&config.static_files_path))
        .fallback(static_files_handler)
        .with_state(config)
}

async fn root_handler(
    axum::extract::State(config): axum::extract::State<ServerConfig>,
) -> axum::response::Response {
    let token = if let Some(sidecar) = config.sidecar {
        sidecar.admin_token.lock().expect("mutex poisoned").clone()
    } else {
        "".to_string()
    };

    let index_path = config.static_files_path.join("index.html");
    let content = if let Ok(c) = tokio::fs::read(&index_path).await {
        axum::response::Html(c).into_response()
    } else {
        let body = "<h1>Pryx Host</h1><p>Local web UI available at /</p>";
        ([("Content-Type", "text/html")], body.to_string()).into_response()
    };

    // Set cookie
    let mut response = content.into_response();
    if !token.is_empty() {
        let cookie = format!(
            "pryx_admin_token={}; Path=/; HttpOnly; SameSite=Strict",
            token
        );
        response
            .headers_mut()
            .insert(header::SET_COOKIE, cookie.parse().unwrap());
    }
    response
}

async fn ws_upgrade_handler(ws: axum::extract::ws::WebSocketUpgrade) -> impl IntoResponse {
    ws.on_upgrade(handle_socket)
}
