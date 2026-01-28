use pryx_host::sidecar::{find_pryx_core_binary, SidecarConfig, SidecarProcess, SidecarStatus};
use pryx_host::sidecar::permissions::{PermissionManager, PermissionDialogConfig, ApprovalRequest, ApprovalResponse};
use std::sync::Arc;
use tauri::{AppHandle, Manager, State};
use tauri_plugin_updater::UpdaterExt;

// Command to get sidecar status
#[tauri::command]
fn get_sidecar_status(state: State<Arc<SidecarProcess>>) -> SidecarStatus {
    state.status()
}

// Command to request tool approval
#[tauri::command]
async fn request_tool_approval(
    app: AppHandle,
    request: ApprovalRequest,
) -> Result<bool, String> {
    let config_path = app.path().join("permissions.json");
    let config = match std::fs::read_to_string(&config_path) {
        Ok(content) => {
            match serde_json::from_str::<PermissionDialogConfig>(&content) {
                Ok(cfg) => cfg,
                Err(_) => PermissionDialogConfig::default(),
            }
        }
        Err(_) => {
            eprintln!("Failed to read permissions config, using defaults");
            PermissionDialogConfig::default()
        }
    };

    let permission_manager = Arc::new(PermissionManager::new(config));
    let manager_clone = permission_manager.clone();
    let request_clone = request.clone();

    let result = tokio::task::spawn_blocking(move || {
        match manager_clone.request_approval(&app, request_clone) {
            Ok(ApprovalResponse::Approved) => true,
            Ok(ApprovalResponse::Denied) => false,
            Ok(ApprovalResponse::Cancelled) => false,
            Err(e) => {
                eprintln!("Approval error: {}", e);
                false
            }
        }
    }).await;

    match result {
        Ok(approved) => Ok(approved),
        Err(e) => Err(e),
    }
}

// Command to list pending approvals
#[tauri::command]
async fn list_pending_approvals(app: AppHandle) -> Vec<ApprovalRequest> {
    let config_path = app.path().join("permissions.json");
    let config = match std::fs::read_to_string(&config_path) {
        Ok(content) => {
            match serde_json::from_str::<PermissionDialogConfig>(&content) {
                Ok(cfg) => cfg,
                Err(_) => PermissionDialogConfig::default(),
            }
        }
        Err(_) => {
            eprintln!("Failed to read permissions config, using defaults");
            PermissionDialogConfig::default()
        }
    };

    let permission_manager = Arc::new(PermissionManager::new(config));
    permission_manager.list_pending()
}

// Command to cancel an approval request
#[tauri::command]
async fn cancel_approval(app: AppHandle, request_id: String) -> Result<(), String> {
    let config_path = app.path().join("permissions.json");
    let config = match std::fs::read_to_string(&config_path) {
        Ok(content) => {
            match serde_json::from_str::<PermissionDialogConfig>(&content) {
                Ok(cfg) => cfg,
                Err(_) => PermissionDialogConfig::default(),
            }
        }
        Err(_) => {
            eprintln!("Failed to read permissions config, using defaults");
            PermissionDialogConfig::default()
        }
    };

    let permission_manager = Arc::new(PermissionManager::new(config));
    permission_manager.cancel_request(request_id);
    eprintln!("Cancelled approval request: {}", request_id);
    Ok(())
}

// Command to configure permissions
#[tauri::command]
async fn configure_permissions(
    app: AppHandle,
    config: PermissionDialogConfig,
) -> Result<(), String> {
    let config_path = app.path().join("permissions.json");
    let config_json = serde_json::to_string_pretty(&config).map_err(|e| e.to_string())?;
    std::fs::write(&config_path, config_json).map_err(|e| e.to_string())?;
    eprintln!("Permissions config updated");
    Ok(())
}

// Command to check if native dialogs are supported
#[tauri::command]
async fn check_native_dialog_support(app: AppHandle) -> bool {
    // macOS and Windows have native dialogs
    #[cfg(target_os = "macos")]
    return true;

    #[cfg(target_os = "windows")]
    return true;

    // Linux may have limited native dialog support depending on desktop environment
    #[cfg(target_os = "linux")]
    return false;
}

#[tauri::command]
async fn show_permission_dialog(
    app: AppHandle,
    message: String,
    title: String,
) -> Result<bool, String> {
    show_permission_dialog(app, message, title).await
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
        update.download_and_install(
            |chunk_length, content_length| {
                downloaded += chunk_length;
                let _ = app.emit("update-progress", serde_json::json!({
                    "downloaded": downloaded,
                    "contentLength": content_length
                }));
            },
            || {
                let _ = app.emit("update-installed", ());
            }
        ).await.map_err(|e| e.to_string())?;
    }
    Ok(())
}

#[tauri::command]
async fn dispatch_notification(
    app: AppHandle,
    title: String,
    body: String,
) -> Result<(), String> {
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
    app.clipboard()
        .read_text()
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn write_clipboard(app: AppHandle, text: String) -> Result<(), String> {
    use tauri_plugin_clipboard_manager::ClipboardExt;
    app.clipboard()
        .write_text(text)
        .map_err(|e| e.to_string())
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
        .plugin(tauri_plugin_updater::Builder::new().build())
        .manage(Arc::new(SidecarProcess::new(SidecarConfig::default())))
        .setup(|app| {
            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                if let Ok(updater) = handle.updater() {
                    if let Ok(Some(update)) = updater.check().await {
                        let _ = handle.emit("update-available", serde_json::json!({
                            "version": update.version,
                            "body": update.body,
                        }));
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
