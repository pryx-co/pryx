use serde_json::Value;
use std::collections::HashMap;
#[cfg(unix)]
use std::os::unix::fs::PermissionsExt;
use std::path::PathBuf;
use std::process::Stdio;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use tauri::AppHandle;
use tauri_plugin_clipboard_manager::ClipboardExt;
use tauri_plugin_dialog::{DialogExt, MessageDialogKind};
use tauri_plugin_notification::NotificationExt;
use tauri_plugin_updater::UpdaterExt;
use tokio::io::{AsyncBufReadExt, AsyncWriteExt, BufReader};
use tokio::process::{Child, ChildStdin, Command};
use tokio::sync::{oneshot, Mutex as AsyncMutex};

use crate::sidecar::config::*;
use crate::sidecar::types::*;

/// Sidecar process information
#[derive(Debug, Clone)]
pub struct SidecarProcess {
    pub(crate) config: SidecarConfig,
    pub(crate) state: Arc<Mutex<SidecarState>>,
    pub(crate) child: Arc<Mutex<Option<Child>>>,
    pub(crate) port: Arc<Mutex<Option<u16>>>,
    pub(crate) start_time: Arc<Mutex<Option<Instant>>>,
    pub(crate) crash_count: Arc<Mutex<u32>>,
    pub(crate) stdin: Arc<AsyncMutex<Option<ChildStdin>>>,
    pub(crate) app_handle: Arc<Mutex<Option<AppHandle>>>,
    pub(crate) admin_token: Arc<Mutex<String>>,
    pub(crate) pending_requests: Arc<AsyncMutex<HashMap<u64, oneshot::Sender<Value>>>>,
    pub(crate) next_rpc_id: Arc<AsyncMutex<u64>>,
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
            admin_token: Arc::new(Mutex::new(generate_random_token(32))),
            pending_requests: Arc::new(AsyncMutex::new(HashMap::new())),
            next_rpc_id: Arc::new(AsyncMutex::new(1)),
        }
    }

    pub fn status(&self) -> SidecarStatus {
        let state = self.state();
        let pid = self
            .child
            .lock()
            .expect("mutex poisoned")
            .as_ref()
            .and_then(|c| c.id());
        let port = self.port();

        let uptime = self
            .start_time
            .lock()
            .expect("mutex poisoned")
            .as_ref()
            .map(|t| t.elapsed().as_secs_f64());

        let crash_count = *self.crash_count.lock().expect("mutex poisoned");
        let started_at = self
            .start_time
            .lock()
            .expect("mutex poisoned")
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
        let state = self.state.lock().expect("mutex poisoned");
        state.clone()
    }

    pub fn port(&self) -> Option<u16> {
        *self.port.lock().expect("mutex poisoned")
    }

    pub async fn start(&self) -> Result<(), SidecarError> {
        log::info!("Starting sidecar: {:?}", self.config.binary);

        {
            let mut state_guard = self.state.lock().expect("mutex poisoned");
            *state_guard = SidecarState::Starting;
            *self.start_time.lock().expect("mutex poisoned") = Some(Instant::now());
        }

        // Save admin token to file with cross-platform support
        let token = self.admin_token.lock().expect("mutex poisoned").clone();
        let token_path = if let Some(home) = dirs::home_dir() {
            home.join(".pryx/admin.token")
        } else {
            PathBuf::from(".pryx/admin.token")
        };

        // Create .pryx directory if it doesn't exist
        if let Some(parent) = token_path.parent() {
            if let Err(e) = std::fs::create_dir_all(parent) {
                log::error!("Failed to create .pryx directory: {}", e);
            }
        }

        // Write token with secure permissions
        if let Err(e) = std::fs::write(&token_path, &token) {
            log::error!("Failed to write admin token to {:?}: {}", token_path, e);
        } else {
            // Set secure permissions on Unix (0o600)
            #[cfg(unix)]
            {
                if let Err(e) =
                    std::fs::set_permissions(&token_path, PermissionsExt::from_mode(0o600))
                {
                    log::error!(
                        "Failed to set secure permissions on {:?}: {}",
                        token_path,
                        e
                    );
                }
            }
        }

        match self.spawn_sidecar().await {
            Ok(child) => {
                *self.child.lock().expect("mutex poisoned") = Some(child);

                let port_result =
                    tokio::time::timeout(self.config.start_timeout, self.wait_for_port()).await;

                match port_result {
                    Ok(Ok(port)) => {
                        *self.state.lock().expect("mutex poisoned") = SidecarState::Running;
                        log::info!("Sidecar started successfully on port {}", port);
                        Ok(())
                    }
                    Ok(Err(e)) => {
                        *self.state.lock().expect("mutex poisoned") = SidecarState::Running;
                        log::warn!("Sidecar started but port discovery failed: {:?}", e);
                        Ok(())
                    }
                    Err(_) => {
                        *self.state.lock().expect("mutex poisoned") = SidecarState::Running;
                        log::warn!("Sidecar started but port discovery timed out");
                        Ok(())
                    }
                }
            }
            Err(e) => {
                *self.state.lock().expect("mutex poisoned") = SidecarState::Stopped;
                Err(e)
            }
        }
    }

    pub async fn stop(&self) -> Result<(), SidecarError> {
        log::info!("Stopping sidecar");
        *self.state.lock().expect("mutex poisoned") = SidecarState::Stopping;

        let child_opt = { self.child.lock().expect("mutex poisoned").take() };

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

        *self.state.lock().expect("mutex poisoned") = SidecarState::Stopped;
        *self.start_time.lock().expect("mutex poisoned") = None;
        *self.port.lock().expect("mutex poisoned") = None;

        Ok(())
    }

    pub async fn monitor(&self) {
        let mut crash_count = 0;

        loop {
            let state = { self.state.lock().expect("mutex poisoned").clone() };

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
                        let mut child_guard = self.child.lock().expect("mutex poisoned");
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
                            let mut child_guard = self.child.lock().expect("mutex poisoned");
                            *child_guard = None;
                        }

                        crash_count += 1;
                        if self.config.max_restarts > 0 && crash_count > self.config.max_restarts {
                            log::error!("Max restarts ({}) exceeded", self.config.max_restarts);
                            *self.state.lock().expect("mutex poisoned") = SidecarState::Crashed {
                                attempts: crash_count,
                            };
                            return;
                        }

                        let backoff = calculate_backoff(crash_count, &self.config);
                        log::info!("Restarting in {}ms (Attempt {})", backoff, crash_count);

                        *self.state.lock().expect("mutex poisoned") = SidecarState::Restarting {
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

    pub async fn call_rpc(&self, method: &str, params: Value) -> anyhow::Result<Value> {
        let id = {
            let mut id_guard = self.next_rpc_id.lock().await;
            let id = *id_guard;
            *id_guard += 1;
            id
        };

        let (tx, rx) = oneshot::channel();
        {
            let mut pending = self.pending_requests.lock().await;
            pending.insert(id, tx);
        }

        let req = serde_json::json!({
            "jsonrpc": "2.0",
            "method": method,
            "params": params,
            "id": id
        });
        let json = serde_json::to_string(&req)?;
        let mut stdin_guard = self.stdin.lock().await;
        if let Some(stdin) = stdin_guard.as_mut() {
            if let Err(e) = stdin.write_all(json.as_bytes()).await {
                let mut pending = self.pending_requests.lock().await;
                pending.remove(&id);
                return Err(anyhow::anyhow!("Failed to write to stdin: {}", e));
            }
            if let Err(e) = stdin.write_all(b"\n").await {
                let mut pending = self.pending_requests.lock().await;
                pending.remove(&id);
                return Err(anyhow::anyhow!("Failed to write newline to stdin: {}", e));
            }
            if let Err(e) = stdin.flush().await {
                let mut pending = self.pending_requests.lock().await;
                pending.remove(&id);
                return Err(anyhow::anyhow!("Failed to flush stdin: {}", e));
            }
        } else {
            let mut pending = self.pending_requests.lock().await;
            pending.remove(&id);
            return Err(anyhow::anyhow!("Sidecar stdin not available"));
        }

        match tokio::time::timeout(Duration::from_secs(10), rx).await {
            Ok(Ok(val)) => {
                // Check if the response contains a JSON-RPC error
                if let Some(error) = val.get("error") {
                    let error_code = error.get("code").and_then(|v| v.as_i64()).unwrap_or(-1);
                    let error_message = error
                        .get("message")
                        .and_then(|v| v.as_str())
                        .unwrap_or("Unknown error");
                    let _error_data = error.get("data");
                    return Err(anyhow::anyhow!(
                        "JSON-RPC error {}: {}",
                        error_code,
                        error_message
                    ));
                }
                Ok(val)
            }
            Ok(Err(_)) => {
                let mut pending = self.pending_requests.lock().await;
                pending.remove(&id);
                Err(anyhow::anyhow!("RPC response channel closed"))
            }
            Err(_) => {
                let mut pending = self.pending_requests.lock().await;
                pending.remove(&id);
                Err(anyhow::anyhow!("RPC request timed out"))
            }
        }
    }

    async fn spawn_sidecar(&self) -> Result<Child, SidecarError> {
        let binary_path = &self.config.binary;

        let mut cmd = Command::new(binary_path);
        cmd.args(&self.config.args);
        cmd.current_dir(&self.config.cwd);
        // Set Envs
        cmd.env("PRYX_LISTEN_ADDR", "127.0.0.1:0");
        cmd.env(
            "PRYX_DB_PATH",
            self.config.db_path.to_string_lossy().to_string(),
        );
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
                        if let Ok(val) = serde_json::from_str::<Value>(&line) {
                            if val.get("result").is_some() || val.get("error").is_some() {
                                // This is a response
                                if let Some(id) = val.get("id").and_then(|v| v.as_u64()) {
                                    let mut pending = process_clone.pending_requests.lock().await;
                                    if let Some(tx) = pending.remove(&id) {
                                        let result =
                                            val.get("result").cloned().unwrap_or(Value::Null);
                                        let _ = tx.send(result);
                                    }
                                }
                            } else if val.get("method").is_some() {
                                // This is a request
                                if let Ok(req) = serde_json::from_value::<RpcRequest>(val) {
                                    log::info!("Received RPC Request: {:?}", req);
                                    let _ = process_clone.handle_rpc(req).await;
                                }
                            }
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
                let app_guard = self.app_handle.lock().expect("mutex poisoned");
                if let Some(app) = app_guard.as_ref() {
                    // Parse params
                    let description = req
                        .params
                        .get("description")
                        .and_then(|v: &Value| v.as_str())
                        .unwrap_or("Unknown Action");
                    let _intent = req
                        .params
                        .get("intent")
                        .and_then(|v: &Value| v.as_str())
                        .unwrap_or("Requested by Runtime");

                    log::info!("Asking permission for: {}", description);

                    app.dialog()
                        .message(description)
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
                let app_guard = self.app_handle.lock().expect("mutex poisoned");
                if let Some(app) = app_guard.as_ref() {
                    let title = req
                        .params
                        .get("title")
                        .and_then(|v: &Value| v.as_str())
                        .unwrap_or("Pryx Notification");
                    let body = req
                        .params
                        .get("body")
                        .and_then(|v: &Value| v.as_str())
                        .unwrap_or("");

                    log::info!("Showing notification: {} - {}", title, body);

                    let _ = app.notification().builder().title(title).body(body).show();

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
            let text = req
                .params
                .get("text")
                .and_then(|v: &Value| v.as_str())
                .unwrap_or("")
                .to_string();
            let success = {
                let app_guard = self.app_handle.lock().expect("mutex poisoned");
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
                let app_guard = self.app_handle.lock().expect("mutex poisoned");
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
            let app = self.app_handle.lock().expect("mutex poisoned").clone();
            if let Some(app) = app {
                log::info!("Checking for updates...");
                let updater_res = app.updater();
                match updater_res {
                    Ok(updater) => match updater.check().await {
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
                    },
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
            let app = self.app_handle.lock().expect("mutex poisoned").clone();
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
                                match update
                                    .download_and_install(
                                        |chunk_length: usize, content_length: Option<u64>| {
                                            if !started {
                                                log::info!(
                                                    "Download started. Total: {:?}",
                                                    content_length
                                                );
                                                started = true;
                                            }
                                            downloaded += chunk_length;
                                            log::debug!("Downloaded {} bytes", downloaded);
                                        },
                                        || {
                                            log::info!("Download finished");
                                        },
                                    )
                                    .await
                                {
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

    pub async fn send_notification(
        &self,
        method: &str,
        params: serde_json::Value,
    ) -> Result<(), SidecarError> {
        let notification = serde_json::json!({
            "jsonrpc": "2.0",
            "method": method,
            "params": params
        });

        let json_line = serde_json::to_string(&notification)
            .map_err(|e| SidecarError::Serialization(e.to_string()))?;

        let mut stdin_guard = self.stdin.lock().await;
        if let Some(stdin) = stdin_guard.as_mut() {
            stdin
                .write_all(json_line.as_bytes())
                .await
                .map_err(SidecarError::Io)?;
            stdin.write_all(b"\n").await.map_err(SidecarError::Io)?;
            stdin.flush().await.map_err(SidecarError::Io)?;
            Ok(())
        } else {
            Err(SidecarError::ProcessNotRunning(
                "Stdin not available".into(),
            ))
        }
    }

    async fn send_response(&self, resp: RpcResponse) -> anyhow::Result<()> {
        let json = serde_json::to_string(&resp)?;
        log::info!("Sending RPC Response: {}", json);

        let mut stdin_guard = self.stdin.lock().await;
        if let Some(stdin) = stdin_guard.as_mut() {
            stdin
                .write_all(json.as_bytes())
                .await
                .map_err(|e| anyhow::anyhow!(e))?;
            stdin
                .write_all(b"\n")
                .await
                .map_err(|e| anyhow::anyhow!(e))?;
            stdin.flush().await.map_err(|e| anyhow::anyhow!(e))?;
        }
        Ok(())
    }

    async fn wait_for_port(&self) -> Result<u16, SidecarError> {
        loop {
            if let Some(port) = *self.port.lock().expect("mutex poisoned") {
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
            let potential = &line[idx + 1..];
            let _ = potential.trim().replace(|c: char| !c.is_numeric(), "");
            let digits: String = potential
                .chars()
                .skip_while(|c| !c.is_numeric())
                .take_while(|c| c.is_numeric())
                .collect();
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

fn generate_random_token(len: usize) -> String {
    use rand::distributions::Alphanumeric;
    use rand::{thread_rng, Rng};

    thread_rng()
        .sample_iter(&Alphanumeric)
        .take(len)
        .map(char::from)
        .collect()
}
