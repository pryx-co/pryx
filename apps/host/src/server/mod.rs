use crate::sidecar::SidecarProcess;
use axum::Router;
use std::net::SocketAddr;
use std::path::PathBuf;
use std::sync::Arc;
use thiserror::Error;

pub mod auth;
pub mod handlers;
pub mod routes;
pub mod websocket;

pub use handlers::{health_handler, skills_handler};
pub use routes::app_router;
pub use websocket::handle_socket;

#[derive(Error, Debug)]
pub enum ServerError {
    #[error("Failed to bind to port: {0}")]
    BindError(std::io::Error),
    #[error("Static file error: {0}")]
    StaticFileError(#[from] std::io::Error),
    #[error("WebSocket error: {0}")]
    WebSocketError(String),
}

#[derive(Clone)]
pub struct ServerConfig {
    pub host: String,
    pub port: u16,
    pub static_files_path: PathBuf,
    pub sidecar: Option<Arc<SidecarProcess>>,
}

impl Default for ServerConfig {
    fn default() -> Self {
        Self {
            host: "127.0.0.1".to_string(),
            port: 42424,
            static_files_path: PathBuf::from("../../local-web/dist"),
            sidecar: None,
        }
    }
}

pub async fn start_server(config: ServerConfig) -> Result<(), ServerError> {
    let addr: SocketAddr = format!("{}:{}", config.host, config.port)
        .parse()
        .map_err(|e| {
            ServerError::BindError(std::io::Error::new(std::io::ErrorKind::InvalidInput, e))
        })?;

    let app: Router = routes::app_router(config.clone());

    log::info!("Starting HTTP server on http://{}", addr);
    log::info!("Serving static files from: {:?}", config.static_files_path);

    let listener = tokio::net::TcpListener::bind(addr)
        .await
        .map_err(ServerError::BindError)?;
    axum::serve(listener, app)
        .await
        .map_err(|e| ServerError::BindError(std::io::Error::other(e)))?;

    Ok(())
}
