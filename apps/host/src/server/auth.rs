use crate::server::ServerConfig;
use axum::{
    body::Body,
    http::{Request, Response, StatusCode},
    middleware::Next,
    response::IntoResponse,
};

pub async fn auth_middleware(
    request: Request<Body>,
    next: Next,
) -> Result<Response<Body>, Response<Body>> {
    let config = request
        .extensions()
        .get::<ServerConfig>()
        .ok_or_else(|| (StatusCode::INTERNAL_SERVER_ERROR, "Config missing").into_response())?;

    // Get sidecar to access the token
    let sidecar = config.sidecar.as_ref().ok_or_else(|| {
        (StatusCode::SERVICE_UNAVAILABLE, "Sidecar not initialized").into_response()
    })?;

    let expected_token = sidecar.admin_token.lock().expect("mutex poisoned").clone();

    // 1. Check Authorization header
    if let Some(auth_header) = request.headers().get("Authorization") {
        if let Ok(auth_str) = auth_header.to_str() {
            if auth_str == format!("Bearer {}", expected_token) {
                return Ok(next.run(request).await);
            }
        }
    }

    // 2. Check Cookie (for browser UI)
    if let Some(cookie_header) = request.headers().get("Cookie") {
        if let Ok(cookie_str) = cookie_header.to_str() {
            if cookie_str.contains(&format!("pryx_admin_token={}", expected_token)) {
                return Ok(next.run(request).await);
            }
        }
    }

    // Unauthorized
    Err((StatusCode::UNAUTHORIZED, "Unauthorized").into_response())
}
