use pryx_host::server::{start_server, ServerConfig};
use pryx_host::sidecar::permissions::*;
use pryx_host::sidecar::*;
use std::sync::Arc;
use tauri::{AppHandle, Emitter, Manager, State};
use tauri_plugin_updater::UpdaterExt;

mod tray;

// Command to get sidecar status
#[tauri::command]
fn get_sidecar_status(state: State<Arc<SidecarProcess>>) -> SidecarStatus {
    state.status()
}

// --- Updater Commands ---

#[tauri::command]
async fn check_for_updates(app: AppHandle) -> Result<bool, String> {
    let updater = app.updater().map_err(|e| e.to_string())?;
    let update = updater.check().await.map_err(|e| e.to_string())?;
    Ok(update.is_some())
}

#[tauri::command]
async fn install_update(app: AppHandle) -> Result<(), String> {
    let updater = app.updater().map_err(|e| e.to_string())?;
    if let Some(update) = updater.check().await.map_err(|e| e.to_string())? {
        let mut downloaded = 0;
        update
            .download_and_install(
                |chunk_length, content_length| {
                    downloaded += chunk_length;
                    let _ = app.emit(
                        "update-progress",
                        serde_json::json!({
                            "downloaded": downloaded,
                            "contentLength": content_length
                        }),
                    );
                },
                || {
                    let _ = app.emit("update-installed", ());
                },
            )
            .await
            .map_err(|e| e.to_string())?;
    }
    Ok(())
}

#[tauri::command]
async fn dispatch_notification(app: AppHandle, title: String, body: String) -> Result<(), String> {
    use tauri_plugin_notification::NotificationExt;
    app.notification()
        .builder()
        .title(title)
        .body(body)
        .show()
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn read_clipboard(app: AppHandle) -> Result<String, String> {
    use tauri_plugin_clipboard_manager::ClipboardExt;
    app.clipboard().read_text().map_err(|e| e.to_string())
}

#[tauri::command]
async fn write_clipboard(app: AppHandle, text: String) -> Result<(), String> {
    use tauri_plugin_clipboard_manager::ClipboardExt;
    app.clipboard().write_text(text).map_err(|e| e.to_string())
}

#[tokio::main]
async fn main() {
    tauri::Builder::default()
        .plugin(tauri_plugin_deep_link::init())
        .plugin(tauri_plugin_shell::init())
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_notification::init())
        .plugin(tauri_plugin_clipboard_manager::init())
        .on_window_event(|window, event| {
            if let tauri::WindowEvent::CloseRequested { api, .. } = event {
                if let Err(e) = window.hide() {
                    log::error!("Failed to hide window on close request: {}", e);
                }
                api.prevent_close();
            }
        })
        .plugin(tauri_plugin_updater::Builder::new().build())
        .setup(|app| {
            // Create and start sidecar process (Go runtime)
            let sidecar_state = Arc::new(SidecarProcess::new(
                SidecarConfig::default(),
                app.handle().clone(),
            ));

            // Start sidecar in background
            let sidecar_clone = sidecar_state.clone();
            tauri::async_runtime::spawn(async move {
                if let Err(e) = sidecar_clone.start().await {
                    log::error!("Failed to start sidecar process: {:?}", e);
                }
            });

            // Start sidecar monitor
            let monitor_clone = sidecar_state.clone();
            tauri::async_runtime::spawn(async move {
                monitor_clone.monitor().await;
            });

            // Manage sidecar state
            app.manage(sidecar_state.clone());

            // Start HTTP server on port 42424
            let sidecar_rpc_clone = sidecar_state;
            let server_config = ServerConfig {
                host: "127.0.0.1".to_string(),
                port: 42424,
                static_files_path: std::path::PathBuf::from("../local-web/dist"),
                sidecar: Some(sidecar_rpc_clone),
            };

            let _server_handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                if let Err(e) = start_server(server_config).await {
                    log::error!("Failed to start HTTP server: {}", e);
                }
            });

            // Deep Link Handler
            #[cfg(any(windows, target_os = "linux"))]
            {
                use tauri_plugin_deep_link::DeepLinkExt;
                app.deep_link().register_all()?;
            }

            // System Tray
            tray::create_tray(app.handle())?;

            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                if let Ok(updater) = handle.updater() {
                    if let Ok(Some(update)) = updater.check().await {
                        let _ = handle.emit(
                            "update-available",
                            serde_json::json!({
                                "version": update.version,
                                "body": update.body,
                            }),
                        );
                    }
                }
            });
            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            get_sidecar_status,
            request_tool_approval,
            list_pending_approvals,
            cancel_approval,
            configure_permissions,
            check_native_dialog_support,
            show_permission_dialog,
            check_for_updates,
            install_update,
            dispatch_notification,
            read_clipboard,
            write_clipboard,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
