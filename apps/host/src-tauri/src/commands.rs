use tauri::AppHandle;
use sidecar::{SidecarProcess, SidecarConfig, SidecarState, SidecarError, SidecarStatus};

#[tauri::command]
pub async fn start_sidecar(app: AppHandle) -> Result<String, String> {
    let config = SidecarConfig::default();
    let sidecar = SidecarProcess::new(config);
    
    if let Err(e) = sidecar.start().await {
        return Err(format!("Failed to start sidecar: {}", e));
    }
    
    let sidecar_clone = sidecar.clone();
    tauri::async_runtime::spawn(async move {
        if let Err(e) = sidecar_clone.monitor().await {
            log::error!("Sidecar monitor error: {}", e);
        }
    });

    Ok("Sidecar started successfully".to_string())
}

#[tauri::command]
pub async fn stop_sidecar(app: AppHandle) -> Result<String, String> {
    let config = SidecarConfig::default();
    let sidecar = SidecarProcess::new(config);
    
    if let Err(e) = sidecar.stop().await {
        return Err(format!("Failed to stop sidecar: {}", e));
    }
    
    Ok("Sidecar stopped successfully".to_string())
}

#[tauri::command]
pub async fn get_sidecar_status(app: AppHandle) -> Result<SidecarStatus, String> {
    let config = SidecarConfig::default();
    let sidecar = SidecarProcess::new(config);
    
    Ok(sidecar.status())
}

#[tauri::command]
pub async fn restart_sidecar(app: AppHandle) -> Result<String, String> {
    let config = SidecarConfig::default();
    let sidecar = SidecarProcess::new(config.clone());
    
    if let Err(e) = sidecar.stop().await {
        return Err(format!("Failed to stop sidecar for restart: {}", e));
    }
    
    tokio::time::sleep(std::time::Duration::from_millis(500)).await;
    
    if let Err(e) = sidecar.start().await {
        return Err(format!("Failed to restart sidecar: {}", e));
    }
    
    let sidecar_clone = sidecar.clone();
    tauri::async_runtime::spawn(async move {
        if let Err(e) = sidecar_clone.monitor().await {
            log::error!("Sidecar monitor error after restart: {}", e);
        }
    });

    Ok("Sidecar restarted successfully".to_string())
}
