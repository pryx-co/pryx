use std::{
    collections::HashMap,
    path::{Path, PathBuf},
    process::Stdio,
    sync::{Arc, Mutex},
    time::{Duration, Instant},
};
use tauri::AppHandle;
use tauri_plugin_dialog::{DialogExt, MessageDialogKind};
use tauri_plugin_notification::NotificationExt;
use tauri_plugin_clipboard_manager::ClipboardExt;
use tauri_plugin_updater::UpdaterExt;
use tokio::sync::Mutex as AsyncMutex;
use tokio::{
    io::{AsyncBufReadExt, AsyncWriteExt, BufReader},
    process::{Child, ChildStdin, Command},
};

use serde::{Deserialize, Serialize};
pub mod permissions;
#[cfg(test)]
mod tests;
use serde_json::Value;

/// Sidecar process state
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SidecarState {
    Stopped,
    Starting,
    Running,
    Crashed { attempts: u32 },
    Restarting { backoff_ms: u64 },
    Stopping,
}

/// Sidecar configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SidecarConfig {
    pub binary: PathBuf,
    pub args: Vec<String>,
    pub env: HashMap<String, String>,
    pub cwd: PathBuf,
    pub db_path: PathBuf,
    pub start_timeout: Duration,
    pub max_restarts: u32,
    pub initial_backoff_ms: u64,
    pub backoff_multiplier: f64,
}

impl SidecarConfig {
    pub fn new(binary: PathBuf, cwd: PathBuf, db_path: PathBuf) -> Self {
        Self {
            binary,
            args: vec![],
            env: HashMap::new(),
            cwd,
            db_path,
            start_timeout: Duration::from_secs(3),
            max_restarts: 10,
            initial_backoff_ms: 1000,
            backoff_multiplier: 2.0,
        }
    }
}

impl Default for SidecarConfig {
    fn default() -> Self {
        Self::new(
            PathBuf::from("pryx-core"),
            std::env::current_dir().unwrap_or_default(),
            PathBuf::from("pryx.db"),
        )
    }
}

/// RPC Request from Sidecar
#[derive(Debug, Deserialize)]
struct RpcRequest {
    #[allow(dead_code)]
    jsonrpc: String,
    method: String,
    params: Value,
    id: u64,
}

/// RPC Response to Sidecar
#[derive(Debug, Serialize)]
struct RpcResponse {
    jsonrpc: String,
    result: Value,
    id: u64,
}

/// Sidecar process information
#[derive(Debug, Clone)]
pub struct SidecarProcess {
    config: SidecarConfig,
    state: Arc<Mutex<SidecarState>>,
    child: Arc<Mutex<Option<Child>>>,
    port: Arc<Mutex<Option<u16>>>,
    start_time: Arc<Mutex<Option<Instant>>>,
    crash_count: Arc<Mutex<u32>>,
    stdin: Arc<AsyncMutex<Option<ChildStdin>>>,
    app_handle: Arc<Mutex<Option<AppHandle>>>,
}

impl SidecarProcess {
    pub fn new(config: SidecarConfig, app_handle: AppHandle) -> Self {
        Self {
            config,
            state: Arc::new(Mutex::new(SidecarState::Stopped)),
            child: Arc::new(Mutex::new(None)),
            port: Arc::new(Mutex::new(None)),
            start_time: Arc::new(Mutex::new(None)),
            crash_count: Arc::new(Mutex::new(0)),
            stdin: Arc::new(AsyncMutex::new(None)),
            app_handle: Arc::new(Mutex::new(Some(app_handle))),
        }
    }

    pub fn status(&self) -> SidecarStatus {
        let state = self.state();
        let pid = self.child.lock().unwrap().as_ref().and_then(|c| c.id());
        let port = self.port();

        let uptime = self
            .start_time
            .lock()
            .unwrap()
            .as_ref()
            .map(|t| t.elapsed().as_secs_f64());

        let crash_count = *self.crash_count.lock().unwrap();
        let started_at = self
            .start_time
            .lock()
            .unwrap()
            .as_ref()
            .map(|t| format!("{:?}", t));

        SidecarStatus {
            state,
            pid,
            port,
            uptime_secs: uptime,
            crash_count,
            started_at,
        }
    }

    pub fn state(&self) -> SidecarState {
        let state = self.state.lock().unwrap();
        state.clone()
    }

    pub fn port(&self) -> Option<u16> {
        *self.port.lock().unwrap()
    }

    pub async fn start(&self) -> Result<(), SidecarError> {
        log::info!("Starting sidecar: {:?}", self.config.binary);

        {
            *self.state.lock().unwrap() = SidecarState::Starting;
            *self.start_time.lock().unwrap() = Some(Instant::now());
        }

        match self.spawn_sidecar().await {
            Ok(child) => {
                *self.child.lock().unwrap() = Some(child);

                let port_result = tokio::time::timeout(
                    self.config.start_timeout,
                    self.wait_for_port(),
                )
                .await;

                match port_result {
                    Ok(Ok(port)) => {
                        *self.state.lock().unwrap() = SidecarState::Running;
                        log::info!("Sidecar started successfully on port {}", port);
                        Ok(())
                    }
                    Ok(Err(e)) => {
                        *self.state.lock().unwrap() = SidecarState::Running;
                        log::warn!("Sidecar started but port discovery failed: {:?}", e);
                        Ok(())
                    }
                    Err(_) => {
                        *self.state.lock().unwrap() = SidecarState::Running;
                        log::warn!("Sidecar started but port discovery timed out");
                        Ok(())
                    }
                }
            }
            Err(e) => {
                *self.state.lock().unwrap() = SidecarState::Stopped;
                Err(e)
            }
        }
    }

    pub async fn stop(&self) -> Result<(), SidecarError> {
        log::info!("Stopping sidecar");
        *self.state.lock().unwrap() = SidecarState::Stopping;

        let child_opt = { self.child.lock().unwrap().take() };

        if let Some(mut child) = child_opt {
            let pid = child.id().unwrap_or_default() as i32;
            log::info!("Sending SIGTERM to sidecar (PID: {:?})", pid);

            #[cfg(unix)]
            unsafe {
                let _ = libc::killpg(pid, libc::SIGTERM);
            }
            #[cfg(not(unix))]
            {
                let _ = child.start_kill();
            }

            let wait_res = tokio::time::timeout(Duration::from_secs(2), child.wait()).await;
            match wait_res {
                Ok(Ok(_)) => {}
                Ok(Err(e)) => return Err(SidecarError::Io(e)),
                Err(_) => {
                    #[cfg(unix)]
                    unsafe {
                        let _ = libc::killpg(pid, libc::SIGKILL);
                    }
                    #[cfg(not(unix))]
                    {
                        let _ = child.start_kill();
                    }
                    let _ = child.wait().await;
                }
            }
        }

        *self.state.lock().unwrap() = SidecarState::Stopped;
        *self.start_time.lock().unwrap() = None;
        *self.port.lock().unwrap() = None;

        Ok(())
    }

    pub async fn monitor(&self) {
        let mut crash_count = 0;

        loop {
            let state = { self.state.lock().unwrap().clone() };

            match state {
                SidecarState::Stopped => {
                    tokio::time::sleep(Duration::from_millis(500)).await;
                }
                SidecarState::Stopping => {
                    tokio::time::sleep(Duration::from_millis(100)).await;
                }
                SidecarState::Running | SidecarState::Starting => {
                    let mut child_dead = false;
                    {
                        let mut child_guard = self.child.lock().unwrap();
                        if let Some(child) = child_guard.as_mut() {
                            match child.try_wait() {
                                Ok(Some(status)) => {
                                    log::warn!("Sidecar exited: {:?}", status);
                                    child_dead = true;
                                }
                                Ok(None) => {}
                                Err(e) => {
                                    log::error!("Error waiting on child: {:?}", e);
                                    child_dead = true;
                                }
                            }
                        } else {
                            child_dead = true;
                        }
                    }

                    if child_dead {
                        {
                            let mut child_guard = self.child.lock().unwrap();
                            *child_guard = None;
                        }

                        crash_count += 1;
                        if self.config.max_restarts > 0 && crash_count > self.config.max_restarts {
                            log::error!("Max restarts ({}) exceeded", self.config.max_restarts);
                            *self.state.lock().unwrap() = SidecarState::Crashed {
                                attempts: crash_count,
                            };
                            return;
                        }

                        let backoff = calculate_backoff(crash_count, &self.config);
                        log::info!("Restarting in {}ms (Attempt {})", backoff, crash_count);

                        *self.state.lock().unwrap() = SidecarState::Restarting {
                            backoff_ms: backoff,
                        };
                        tokio::time::sleep(Duration::from_millis(backoff)).await;

                        if let Err(e) = self.start().await {
                            log::error!("Failed to restart sidecar: {:?}", e);
                        }
                    } else {
                        tokio::time::sleep(Duration::from_secs(1)).await;
                    }
                }
                SidecarState::Restarting { .. } => {
                    tokio::time::sleep(Duration::from_millis(100)).await;
                }
                SidecarState::Crashed { .. } => {
                    tokio::time::sleep(Duration::from_secs(1)).await;
                }
            }
        }
    }

    pub async fn shutdown(&self) {
        let _ = self.stop().await;
    }

    async fn spawn_sidecar(&self) -> Result<Child, SidecarError> {
        let binary_path = &self.config.binary;

        let mut cmd = Command::new(binary_path);
        cmd.args(&self.config.args);
        cmd.current_dir(&self.config.cwd);
        // Set Envs
        cmd.env("PRYX_LISTEN_ADDR", "127.0.0.1:0");
        cmd.env("PRYX_DB_PATH", self.config.db_path.to_string_lossy().to_string());
        cmd.env("PRYX_HOST_RPC", "1");
        for (k, v) in &self.config.env {
            cmd.env(k, v);
        }
        cmd.stdout(Stdio::piped());
        cmd.stderr(Stdio::piped());
        cmd.stdin(Stdio::piped());

        #[cfg(unix)]
        {
            unsafe {
                cmd.pre_exec(|| {
                    if libc::setpgid(0, 0) != 0 {
                        return Err(std::io::Error::last_os_error());
                    }
                    Ok(())
                });
            }
        }

        let mut child = cmd.spawn().map_err(|e| SidecarError::SpawnFailed {
            binary: binary_path.to_string_lossy().to_string(),
            reason: e.to_string(),
        })?;

        log::info!("Spawned sidecar process (PID: {:?})", child.id());

        // Capture stdin
        if let Some(stdin) = child.stdin.take() {
            *self.stdin.lock().await = Some(stdin);
        }

        // Capture stdout for port discovery AND RPC
        if let Some(stdout) = child.stdout.take() {
            let reader = BufReader::new(stdout);
            let port = self.port.clone();
            let process_clone = self.clone();

            tokio::spawn(async move {
                let mut lines = reader.lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    // 1. Try generic log
                    log::info!("[SIDECAR] {}", line);

                    // 2. Check for port
                    if extract_port_check(&line) {
                        if let Some(p) = extract_port_from_line(&line) {
                            if let Ok(mut port_guard) = port.lock() {
                                if port_guard.is_none() {
                                    *port_guard = Some(p);
                                    log::info!("Discovered sidecar port: {}", p);
                                }
                            }
                        }
                    }

                    // 3. Check for RPC
                    if line.trim().starts_with('{') {
                         if let Ok(req) = serde_json::from_str::<RpcRequest>(&line) {
                             log::info!("Received RPC Request: {:?}", req);
                             let _ = process_clone.handle_rpc(req).await;
                         }
                    }
                }
            });
        }

        // Stderr logging
        if let Some(stderr) = child.stderr.take() {
            let reader = BufReader::new(stderr);
            tokio::spawn(async move {
                let mut lines = reader.lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    log::error!("[SIDECAR ERR] {}", line);
                }
            });
        }

        Ok(child)
    }

    async fn handle_rpc(&self, req: RpcRequest) -> anyhow::Result<()> {
        if req.method == "permission.request" {
            let ans = {
                // Check app handle
                let app_guard = self.app_handle.lock().unwrap();
                if let Some(app) = app_guard.as_ref() {
                    // Parse params
                    let description = req.params.get("description").and_then(|v: &Value| v.as_str()).unwrap_or("Unknown Action");
                    let _intent = req.params.get("intent").and_then(|v: &Value| v.as_str()).unwrap_or("Requested by Runtime");

                    log::info!("Asking permission for: {}", description);

                    app.dialog().message(description)
                        .title("Permission Request")
                        .kind(MessageDialogKind::Warning)
                        .blocking_show()
                } else {
                    log::error!("Cannot handle RPC: No AppHandle");
                    return Ok(());
                }
            };
            
            // Construct response
            let result = serde_json::json!({
                "approved": ans
            });
            
            let resp = RpcResponse {
                jsonrpc: "2.0".to_string(),
                result,
                id: req.id,
            };
            
            self.send_response(resp).await?;
        } else if req.method == "notification.show" {
             let notification_result = {
                 // Check app handle
                 let app_guard = self.app_handle.lock().unwrap();
                 if let Some(app) = app_guard.as_ref() {
                     let title = req.params.get("title").and_then(|v: &Value| v.as_str()).unwrap_or("Pryx Notification");
                     let body = req.params.get("body").and_then(|v: &Value| v.as_str()).unwrap_or("");

                     log::info!("Showing notification: {} - {}", title, body);
                     
                     let _ = app.notification()
                        .builder()
                        .title(title)
                        .body(body)
                        .show();
                     
                     true
                 } else {
                     false
                 }
             };

             if notification_result {
                  // Fire and forget response or simple ack
                 let resp = RpcResponse {
                    jsonrpc: "2.0".to_string(),
                    result: serde_json::json!({"status": "ok"}),
                    id: req.id,
                };
                self.send_response(resp).await?;
             }
         } else if req.method == "clipboard.writeText" {
             let text = req.params.get("text").and_then(|v: &Value| v.as_str()).unwrap_or("").to_string();
             let success = {
                let app_guard = self.app_handle.lock().unwrap();
                if let Some(app) = app_guard.as_ref() {
                    let _ = app.clipboard().write_text(text);
                    true
                } else {
                    false
                }
             };

             if success {
                 let resp = RpcResponse {
                    jsonrpc: "2.0".to_string(),
                    result: serde_json::json!({"status": "ok"}),
                    id: req.id,
                };
                self.send_response(resp).await?;
             }
         } else if req.method == "clipboard.readText" {
             let content = {
                let app_guard = self.app_handle.lock().unwrap();
                if let Some(app) = app_guard.as_ref() {
                    app.clipboard().read_text().unwrap_or_default()
                } else {
                    String::new()
                }
             };
             
             let resp = RpcResponse {
                jsonrpc: "2.0".to_string(),
                result: serde_json::json!({"text": content}),
                id: req.id,
            };
            self.send_response(resp).await?;
         } else if req.method == "updater.check" {
             let app = self.app_handle.lock().unwrap().clone();
             if let Some(app) = app {
                 log::info!("Checking for updates...");
                 let updater_res = app.updater();
                 match updater_res {
                     Ok(updater) => {
                         match updater.check().await {
                             Ok(Some(update)) => {
                                 let body = update.body.clone().unwrap_or_default();
                                 let version = update.version.clone();
                                 log::info!("Update found: {} - {}", version, body);
                                 
                                 let resp = RpcResponse {
                                     jsonrpc: "2.0".to_string(),
                                     result: serde_json::json!({
                                         "available": true,
                                         "version": version,
                                         "body": body
                                     }),
                                     id: req.id,
                                 };
                                 self.send_response(resp).await?;
                             }
                             Ok(None) => {
                                 log::info!("No updates available");
                                 let resp = RpcResponse {
                                     jsonrpc: "2.0".to_string(),
                                     result: serde_json::json!({"available": false}),
                                     id: req.id,
                                 };
                                 self.send_response(resp).await?;
                             }
                             Err(e) => {
                                 log::error!("Update check failed: {}", e);
                                 let resp = RpcResponse {
                                     jsonrpc: "2.0".to_string(),
                                     result: serde_json::json!({"error": e.to_string()}),
                                     id: req.id,
                                 };
                                 self.send_response(resp).await?;
                             }
                         }
                     }
                     Err(e) => {
                         log::error!("Failed to initialize updater: {}", e);
                         let resp = RpcResponse {
                             jsonrpc: "2.0".to_string(),
                             result: serde_json::json!({"error": e.to_string()}),
                             id: req.id,
                         };
                         self.send_response(resp).await?;
                     }
                 }
             }
         } else if req.method == "updater.install" {
             let app = self.app_handle.lock().unwrap().clone();
             if let Some(app) = app {
                 log::info!("Installing update...");
                 // Re-check to get the update object (stateless RPC)
                 let updater_res = app.updater();
                 match updater_res {
                     Ok(updater) => {
                        match updater.check().await {
                             Ok(Some(update)) => {
                                 let mut downloaded = 0;
                                 let mut started = false;
                                 
                                 // We can iterate over events if needed, but for now just download and install
                                 match update.download_and_install(|chunk_length: usize, content_length: Option<u64>| {
                                     if !started {
                                         log::info!("Download started. Total: {:?}", content_length);
                                         started = true;
                                     }
                                     downloaded += chunk_length;
                                     log::debug!("Downloaded {} bytes", downloaded);
                                 }, || {
                                     log::info!("Download finished");
                                 }).await {
                                     Ok(_) => {
                                         log::info!("Update installed. Restarting...");
                                         let resp = RpcResponse {
                                             jsonrpc: "2.0".to_string(),
                                             result: serde_json::json!({"status": "installed", "restart": true}),
                                             id: req.id,
                                         };
                                         self.send_response(resp).await?;
                                         app.restart();
                                     }
                                     Err(e) => {
                                         log::error!("Install failed: {}", e);
                                          let resp = RpcResponse {
                                             jsonrpc: "2.0".to_string(),
                                             result: serde_json::json!({"error": e.to_string()}),
                                             id: req.id,
                                         };
                                         self.send_response(resp).await?;
                                     }
                                 }
                             }
                             _ => {
                                 let resp = RpcResponse {
                                     jsonrpc: "2.0".to_string(),
                                     result: serde_json::json!({"error": "No update found to install"}),
                                     id: req.id,
                                 };
                                 self.send_response(resp).await?;
                             }
                         }
                     }
                     Err(e) => {
                         let resp = RpcResponse {
                             jsonrpc: "2.0".to_string(),
                             result: serde_json::json!({"error": e.to_string()}),
                             id: req.id,
                         };
                         self.send_response(resp).await?;
                     }
                 }
             }
         }
        Ok(())
    }

    pub async fn send_notification(&self, method: &str, params: serde_json::Value) -> Result<(), SidecarError> {
         let notification = serde_json::json!({
             "jsonrpc": "2.0",
             "method": method,
             "params": params
         });

         let json_line = serde_json::to_string(&notification).map_err(|e| SidecarError::Serialization(e.to_string()))?;

         let mut stdin_guard = self.stdin.lock().await;
         if let Some(stdin) = stdin_guard.as_mut() {
             stdin.write_all(json_line.as_bytes()).await.map_err(SidecarError::Io)?;
             stdin.write_all(b"\n").await.map_err(SidecarError::Io)?;
             stdin.flush().await.map_err(SidecarError::Io)?;
             Ok(())
         } else {
             Err(SidecarError::ProcessNotRunning("Stdin not available".into()))
         }
    }

    async fn send_response(&self, resp: RpcResponse) -> anyhow::Result<()> {
        let json = serde_json::to_string(&resp)?;
        log::info!("Sending RPC Response: {}", json);
        
        let mut stdin_guard = self.stdin.lock().await;
        if let Some(stdin) = stdin_guard.as_mut() {
            stdin.write_all(json.as_bytes()).await.map_err(|e| anyhow::anyhow!(e))?;
            stdin.write_all(b"\n").await.map_err(|e| anyhow::anyhow!(e))?;
            stdin.flush().await.map_err(|e| anyhow::anyhow!(e))?;
        }
        Ok(())
    }

    async fn wait_for_port(&self) -> Result<u16, SidecarError> {
        loop {
            if let Some(port) = *self.port.lock().unwrap() {
                return Ok(port);
            }
            tokio::time::sleep(Duration::from_millis(50)).await;
        }
    }
}

// Helpers
fn extract_port_check(line: &str) -> bool {
    line.starts_with("PRYX_CORE_LISTEN_ADDR=") || {
        let lower = line.to_lowercase();
        lower.contains("listening") || lower.contains("port")
    }
}

fn extract_port_from_line(line: &str) -> Option<u16> {
    if let Some(rest) = line.strip_prefix("PRYX_CORE_LISTEN_ADDR=") {
        let v = rest.trim();
        if let Some(port_str) = v.rsplit(':').next() {
            if let Ok(p) = port_str.parse::<u16>() {
                return Some(p);
            }
        }
    }
    if let Some(idx) = line.rfind(':') {
        if idx + 1 < line.len() {
             let potential = &line[idx+1..];
             let _ = potential.trim().replace(|c: char| !c.is_numeric(), "");
             let digits: String = potential.chars().skip_while(|c| !c.is_numeric()).take_while(|c| c.is_numeric()).collect();
             if !digits.is_empty() {
                 return digits.parse().ok();
             }
        }
    }
    None
}

fn calculate_backoff(attempt: u32, config: &SidecarConfig) -> u64 {
    let base = config.initial_backoff_ms as f64;
    let multiplier = config.backoff_multiplier;
    let p = (attempt as i32 - 1).clamp(0, 10);
    let backoff = base * multiplier.powi(p);
    backoff as u64
}

pub fn find_pryx_core_binary() -> Option<PathBuf> {
    if let Ok(p) = std::env::var("PRYX_CORE_PATH") {
        let p = PathBuf::from(p);
        if p.exists() {
            return Some(p);
        }
    }

    if let Ok(exe) = std::env::current_exe() {
        if let Some(p) = search_ancestors(&exe) {
            return Some(p);
        }
    }
    if let Ok(cwd) = std::env::current_dir() {
        if let Some(p) = search_ancestors(&cwd) {
            return Some(p);
        }
    }

    Some(PathBuf::from("pryx-core"))
}

fn search_ancestors(start: &Path) -> Option<PathBuf> {
    for a in start.ancestors().take(8) {
        let c = a.join("apps").join("runtime").join("pryx-core");
        if c.exists() {
            return Some(c);
        }
        // Check dist/bin
        let d = a.join("dist").join("pryx-core");
         if d.exists() {
            return Some(d);
        }
    }
    None
}

#[derive(Debug, thiserror::Error)]
pub enum SidecarError {
    #[error("Failed to spawn sidecar binary '{binary}': {reason}")]
    SpawnFailed { binary: String, reason: String },

    #[error("Sidecar process not running")]
    NoChild,

    #[error("Sidecar process not running: {0}")]
    ProcessNotRunning(String),

    #[error("Port discovery failed: {0}")]
    PortDiscoveryFailed(std::io::Error),

    #[error("Serialization error: {0}")]
    Serialization(String),

    #[error("IO error: {0}")]
    Io(#[from] std::io::Error),

    #[error("Operation cancelled")]
    Cancelled,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SidecarStatus {
    pub state: SidecarState,
    pub pid: Option<u32>,
    pub port: Option<u16>,
    pub uptime_secs: Option<f64>,
    pub crash_count: u32,
    pub started_at: Option<String>,
}
