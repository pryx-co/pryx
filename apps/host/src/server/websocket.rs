use axum::extract::ws::WebSocket;
use futures::StreamExt;
use std::sync::Arc;
use tokio::sync::broadcast;

pub struct WsState {
    pub tx: broadcast::Sender<String>,
}

pub async fn handle_socket(mut socket: WebSocket) {
    while let Some(result) = socket.next().await {
        match result {
            Ok(msg) => {
                if let Ok(text) = msg.to_text() {
                    println!("WebSocket received: {}", text);
                    // Echo back
                    let _ = socket.send(msg).await;
                }
            }
            Err(e) => {
                eprintln!("WebSocket error: {}", e);
                break;
            }
        }
    }
}

pub fn broadcast_message(state: &Arc<WsState>, message: &str) {
    let _ = state.tx.send(message.to_string());
}
