use crate::server::ServerConfig;
use axum::{
    extract::{Path, State},
    http::StatusCode,
    response::{IntoResponse, Response},
    Json,
};
use serde_json::Value;

// Helper to call RPC and handle response
async fn rpc_call(config: ServerConfig, method: &str, params: Value) -> Response {
    if let Some(sidecar) = config.sidecar {
        match sidecar.call_rpc(method, params).await {
            Ok(res) => Json(res).into_response(),
            Err(e) => {
                log::error!("RPC Error ({}): {}", method, e);
                (StatusCode::INTERNAL_SERVER_ERROR, e.to_string()).into_response()
            }
        }
    } else {
        (StatusCode::SERVICE_UNAVAILABLE, "Sidecar not initialized").into_response()
    }
}

pub async fn health_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.health", Value::Null).await
}

pub async fn skills_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.skills.list", Value::Null).await
}

pub async fn config_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.config.get", Value::Null).await
}

pub async fn providers_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.providers.list", Value::Null).await
}

pub async fn models_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.models.list", Value::Null).await
}

// --- Channel Handlers ---

pub async fn channels_list_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.channels.list", Value::Null).await
}

pub async fn channel_create_handler(
    State(config): State<ServerConfig>,
    Json(body): Json<Value>,
) -> Response {
    rpc_call(config, "admin.channels.create", body).await
}

pub async fn channel_delete_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(
        config,
        "admin.channels.delete",
        serde_json::json!({ "id": id }),
    )
    .await
}

pub async fn channel_test_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(
        config,
        "admin.channels.test",
        serde_json::json!({ "id": id }),
    )
    .await
}

pub async fn channel_get_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(
        config,
        "admin.channels.get",
        serde_json::json!({ "id": id }),
    )
    .await
}

pub async fn channel_update_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
    Json(body): Json<Value>,
) -> Response {
    let mut params = body;
    if let Some(obj) = params.as_object_mut() {
        obj.insert("id".to_string(), Value::String(id));
    }
    rpc_call(config, "admin.channels.update", params).await
}

// --- MCP Handlers ---

pub async fn mcp_list_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.mcp.list", Value::Null).await
}

pub async fn mcp_create_handler(
    State(config): State<ServerConfig>,
    Json(body): Json<Value>,
) -> Response {
    rpc_call(config, "admin.mcp.create", body).await
}

pub async fn mcp_delete_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(config, "admin.mcp.delete", serde_json::json!({ "id": id })).await
}

pub async fn mcp_get_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(config, "admin.mcp.get", serde_json::json!({ "id": id })).await
}

pub async fn mcp_update_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
    Json(body): Json<Value>,
) -> Response {
    let mut params = body;
    if let Some(obj) = params.as_object_mut() {
        obj.insert("id".to_string(), Value::String(id));
    }
    rpc_call(config, "admin.mcp.update", params).await
}

// --- Policy Handlers ---

pub async fn policy_list_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.policies.list", Value::Null).await
}

pub async fn policy_create_handler(
    State(config): State<ServerConfig>,
    Json(body): Json<Value>,
) -> Response {
    rpc_call(config, "admin.policies.create", body).await
}

pub async fn policy_get_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(
        config,
        "admin.policies.get",
        serde_json::json!({ "id": id }),
    )
    .await
}

pub async fn policy_update_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
    Json(body): Json<Value>,
) -> Response {
    let mut params = body;
    if let Some(obj) = params.as_object_mut() {
        obj.insert("id".to_string(), Value::String(id));
    }
    rpc_call(config, "admin.policies.update", params).await
}

pub async fn policy_delete_handler(
    State(config): State<ServerConfig>,
    Path(id): Path<String>,
) -> Response {
    rpc_call(
        config,
        "admin.policies.delete",
        serde_json::json!({ "id": id }),
    )
    .await
}

// --- Audit Handlers ---

pub async fn audit_list_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.audit.list", Value::Null).await
}

// --- Cost Handlers ---

pub async fn cost_summary_handler(State(config): State<ServerConfig>) -> Response {
    rpc_call(config, "admin.cost.summary", Value::Null).await
}
