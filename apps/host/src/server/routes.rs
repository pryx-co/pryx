use std::path::PathBuf;
use axum::{
    routing::{get, post, delete},
    Router,
    response::IntoResponse,
    http::{StatusCode, header},
    middleware,
};
use tower_http::services::ServeDir;
use crate::server::ServerConfig;
use super::handlers::{
    health_handler, skills_handler, config_handler, providers_handler, models_handler,
    channels_list_handler, channel_create_handler, channel_delete_handler, channel_test_handler,
    channel_get_handler, channel_update_handler,
    mcp_list_handler, mcp_create_handler, mcp_delete_handler,
    policy_list_handler, policy_create_handler, policy_get_handler, policy_update_handler, policy_delete_handler,
    audit_list_handler, cost_summary_handler
};
use super::websocket::handle_socket;
use super::auth::auth_middleware;

async fn static_files_handler(uri: axum::http::Uri) -> axum::response::Response {
    let path = uri.path().trim_start_matches('/');

    if path.is_empty() || path == "index.html" {
        let index_path = PathBuf::from("../../local-web/dist/index.html");
        if let Ok(content) = tokio::fs::read(index_path).await {
            return axum::response::Html(content).into_response();
        }
    }

    let file_path = PathBuf::from("../../local-web/dist").join(path);

    if file_path.exists() && file_path.is_file() {
        if let Ok(content) = tokio::fs::read(&file_path).await {
            let mime_type = mime_guess::from_path(&file_path).first_or_octet_stream();
            return ([("Content-Type", mime_type.as_ref())], content).into_response();
        }
    }

    (StatusCode::NOT_FOUND, "File not found").into_response()
}

pub fn app_router(config: ServerConfig) -> Router {
    let api_routes = Router::new()
        .route("/health", get(health_handler))
        .route("/skills", get(skills_handler))
        .route("/config", get(config_handler))
        .route("/providers", get(providers_handler))
        .route("/models", get(models_handler))
        // Channels
        .route("/channels", get(channels_list_handler).post(channel_create_handler))
        .route("/channels/:id", get(channel_get_handler).put(channel_update_handler).delete(channel_delete_handler))
        .route("/channels/:id/test", post(channel_test_handler))
        // MCP
        .route("/mcp", get(mcp_list_handler).post(mcp_create_handler))
        .route("/mcp/:id", delete(mcp_delete_handler))
        // Policies
        .route("/policies", get(policy_list_handler).post(policy_create_handler))
        .route("/policies/:id", get(policy_get_handler).put(policy_update_handler).delete(policy_delete_handler))
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

async fn root_handler(axum::extract::State(config): axum::extract::State<ServerConfig>) -> axum::response::Response {
    let token = if let Some(sidecar) = config.sidecar {
        sidecar.admin_token.lock().expect("mutex poisoned").clone()
    } else {
        "".to_string()
    };

    let index_path = PathBuf::from("../../local-web/dist/index.html");
    let content = if let Ok(c) = tokio::fs::read(index_path).await {
        axum::response::Html(c).into_response()
    } else {
        let body = "<h1>Pryx Host</h1><p>Local web UI available at /</p>";
        ([("Content-Type", "text/html")], body.to_string()).into_response()
    };

    // Set cookie
    let mut response = content.into_response();
    if !token.is_empty() {
        let cookie = format!("pryx_admin_token={}; Path=/; HttpOnly; SameSite=Strict", token);
        response.headers_mut().insert(header::SET_COOKIE, cookie.parse().unwrap());
    }
    response
}

async fn ws_upgrade_handler(ws: axum::extract::ws::WebSocketUpgrade) -> impl IntoResponse {
    ws.on_upgrade(handle_socket)
}