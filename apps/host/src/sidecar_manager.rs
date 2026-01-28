use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::process::Stdio;
use std::sync::atomic::{AtomicBool, AtomicI32, AtomicU64, Ordering};
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};

use anyhow::{anyhow, Result};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use tokio::io::AsyncBufReadExt;
use tokio::io::AsyncWriteExt;
use tokio::process::{Child, ChildStdin, Command};
use tokio::sync::Mutex as AsyncMutex;

use tauri::AppHandle;
use tauri_plugin_dialog::{DialogExt, MessageDialogKind};

#[derive(Clone, Debug, Deserialize)]
struct RpcRequest {
    jsonrpc: String,
    method: String,
    params: HashMap<String, Value>,
    id: u64,
}

#[derive(Clone, Debug, Serialize)]
struct RpcResponse {
    jsonrpc: String,
    result: Value,
    id: u64,
}

#[derive(Clone, Debug)]
pub struct SidecarConfig {
    pub binary: PathBuf,
    pub args: Vec<String>,
    pub env: HashMap<String, String>,
    pub cwd: PathBuf,
    pub db_path: PathBuf,
}

impl SidecarConfig {
    pub fn new(binary: PathBuf, cwd: PathBuf, db_path: PathBuf) -> Self {
        Self {
            binary,
            args: vec![],
            env: HashMap::new(),
            cwd,
            db_path,
        }
    }
}

#[derive(Clone, Debug, Serialize, Deserialize)]
pub struct SidecarStatus {
    pub running: bool,
    pub pid: Option<u32>,
    pub listen_addr: Option<String>,
    pub restarts: u64,
    pub last_exit_code: Option<i32>,
    pub last_error: Option<String>,
    pub started_at_ms: Option<u128>,
}

impl Default for SidecarStatus {
    fn default() -> Self {
        Self {
            running: false,
            pid: None,
            listen_addr: None,
            restarts: 0,
            last_exit_code: None,
            last_error: None,
            started_at_ms: None,
        }
    }
}

#[derive(Clone)]
pub struct SidecarProcess {
    config: SidecarConfig,
    child: Arc<Mutex<Option<Child>>>,
    stdin: Arc<AsyncMutex<Option<ChildStdin>>>,
    status: Arc<Mutex<SidecarStatus>>,
    shutting_down: Arc<AtomicBool>,
    pgid: Arc<AtomicI32>,
    restarts: Arc<AtomicU64>,
    started_at: Arc<Mutex<Option<Instant>>>,
    app_handle: Arc<Mutex<AppHandle>>,
}

impl SidecarProcess {
    pub fn new(config: SidecarConfig, app_handle: AppHandle) -> Self {
        Self {
            config,
            child: Arc::new(Mutex::new(None)),
            stdin: Arc::new(AsyncMutex::new(None)),
            status: Arc::new(Mutex::new(SidecarStatus::default())),
            shutting_down: Arc::new(AtomicBool::new(false)),
            pgid: Arc::new(AtomicI32::new(0)),
            restarts: Arc::new(AtomicU64::new(0)),
            started_at: Arc::new(Mutex::new(None)),
            app_handle: Arc::new(Mutex::new(app_handle)),
        }
    }

    pub fn status(&self) -> SidecarStatus {
        let mut out = self.status.lock().unwrap().clone();
        out.pid = self.child.lock().unwrap().as_ref().and_then(|c| c.id());
        out.restarts = self.restarts.load(Ordering::SeqCst);
        out.started_at_ms = self
            .started_at
            .lock()
            .unwrap()
            .as_ref()
            .map(|t| t.elapsed().as_millis());
        out
    }

    pub async fn start(&self) -> Result<()> {
        if self.child.lock().unwrap().is_some() {
            return Ok(());
        }

        let mut status = self.status.lock().unwrap();
        status.running = false;
        status.last_error = None;
        status.last_exit_code = None;
        status.listen_addr = None;
        drop(status);

        let mut cmd = Command::new(resolve_binary(&self.config.binary));
        cmd.current_dir(&self.config.cwd);
        cmd.args(&self.config.args);
        cmd.env("PRYX_DB_PATH", &self.config.db_path);
        cmd.env("PRYX_LISTEN_ADDR", "127.0.0.1:0");
        cmd.env("PRYX_HOST_RPC", "1");
        for (k, v) in &self.config.env {
            cmd.env(k, v);
        }
        cmd.stdin(Stdio::piped());
        cmd.stdout(Stdio::piped());
        cmd.stderr(Stdio::piped());

        #[cfg(unix)]
        unsafe {
            cmd.pre_exec(|| {
                libc::setpgid(0, 0);
                Ok(())
            });
        }

        let mut child = cmd.spawn().map_err(|e| anyhow!(e.to_string()))?;
        let pid = child.id().unwrap_or(0) as i32;
        if pid != 0 {
            self.pgid.store(pid, Ordering::SeqCst);
        }

        let stdout = child.stdout.take();
        let stderr = child.stderr.take();
        {
            let mut stdin_guard = self.stdin.lock().await;
            *stdin_guard = child.stdin.take();
        }

        *self.started_at.lock().unwrap() = Some(Instant::now());
        *self.child.lock().unwrap() = Some(child);

        if let Some(stdout) = stdout {
            let status = self.status.clone();
            let stdin = self.stdin.clone();
            let app_handle = self.app_handle.clone();
            tauri::async_runtime::spawn(async move {
                let mut lines = tokio::io::BufReader::new(stdout).lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    if let Some(addr) = parse_listen_addr(&line) {
                        let mut st = status.lock().unwrap();
                        st.listen_addr = Some(addr);
                        continue;
                    }

                    if let Some(req) = parse_permission_request(&line) {
                        let approved = {
                            let app = app_handle.lock().unwrap().clone();
                            let desc = req
                                .params
                                .get("description")
                                .and_then(|v| v.as_str())
                                .unwrap_or("Allow this action?");
                            let intent = req
                                .params
                                .get("intent")
                                .and_then(|v| v.as_str())
                                .unwrap_or("");
                            let msg = if intent.is_empty() {
                                desc.to_string()
                            } else {
                                format!("{}\n\n{}", desc, intent)
                            };
                            app.dialog()
                                .message(msg)
                                .title("Permission Request")
                                .kind(MessageDialogKind::Warning)
                                .blocking_show()
                        };

                        let resp = RpcResponse {
                            jsonrpc: "2.0".to_string(),
                            result: serde_json::json!({ "approved": approved }),
                            id: req.id,
                        };

                        if let Ok(b) = serde_json::to_vec(&resp) {
                            let mut guard = stdin.lock().await;
                            if let Some(stdin) = guard.as_mut() {
                                let _ = stdin.write_all(&b).await;
                                let _ = stdin.write_all(b"\n").await;
                                let _ = stdin.flush().await;
                            }
                        }
                    }
                }
            });
        }
        if let Some(stderr) = stderr {
            let status = self.status.clone();
            tauri::async_runtime::spawn(async move {
                let mut lines = tokio::io::BufReader::new(stderr).lines();
                while let Ok(Some(line)) = lines.next_line().await {
                    let mut st = status.lock().unwrap();
                    st.last_error = Some(line);
                }
            });
        }

        let started_at = Instant::now();
        while started_at.elapsed() < Duration::from_secs(1) {
            let has_addr = { self.status.lock().unwrap().listen_addr.is_some() };
            if has_addr {
                break;
            }
            tokio::time::sleep(Duration::from_millis(25)).await;
        }

        let mut st = self.status.lock().unwrap();
        st.running = true;
        st.pid = self.child.lock().unwrap().as_ref().and_then(|c| c.id());
        drop(st);

        Ok(())
    }

    pub async fn monitor(&self) {
        loop {
            if self.shutting_down.load(Ordering::SeqCst) {
                return;
            }

            let exit_status = {
                let mut guard = self.child.lock().unwrap();
                if let Some(child) = guard.as_mut() {
                    child.try_wait().ok().flatten()
                } else {
                    None
                }
            };

            if let Some(status) = exit_status {
                *self.child.lock().unwrap() = None;
                self.pgid.store(0, Ordering::SeqCst);

                self.restarts.fetch_add(1, Ordering::SeqCst);
                let attempts = self.restarts.load(Ordering::SeqCst) as u32;

                {
                    let mut st = self.status.lock().unwrap();
                    st.running = false;
                    st.pid = None;
                    st.last_exit_code = status.code();
                }

                tokio::time::sleep(restart_backoff(attempts)).await;
                let _ = self.start().await;
            } else {
                tokio::time::sleep(Duration::from_millis(200)).await;
            }
        }
    }

    pub async fn shutdown(&self) {
        self.shutting_down.store(true, Ordering::SeqCst);

        let pgid = self.pgid.load(Ordering::SeqCst);
        if pgid != 0 {
            terminate_process_group(pgid);
        }

        if let Some(mut child) = self.child.lock().unwrap().take() {
            let wait = tokio::time::timeout(Duration::from_millis(2_000), child.wait()).await;
            if wait.is_err() {
                if pgid != 0 {
                    kill_process_group(pgid);
                } else {
                    let _ = child.start_kill();
                }
                let _ = child.wait().await;
            }
        }

        self.pgid.store(0, Ordering::SeqCst);
        *self.started_at.lock().unwrap() = None;

        let mut st = self.status.lock().unwrap();
        st.running = false;
        st.pid = None;
    }
}

#[cfg(unix)]
fn terminate_process_group(pgid: i32) {
    unsafe {
        let _ = libc::killpg(pgid, libc::SIGTERM);
    }
}

#[cfg(unix)]
fn kill_process_group(pgid: i32) {
    unsafe {
        let _ = libc::killpg(pgid, libc::SIGKILL);
    }
}

#[cfg(not(unix))]
fn terminate_process_group(_pgid: i32) {}

#[cfg(not(unix))]
fn kill_process_group(_pgid: i32) {}

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

    None
}

fn search_ancestors(start: &Path) -> Option<PathBuf> {
    for a in start.ancestors().take(8) {
        let c = a.join("apps").join("runtime").join("pryx-core");
        if c.exists() {
            return Some(c);
        }
        let c = a.join("apps").join("runtime").join("bin").join("pryx-core");
        if c.exists() {
            return Some(c);
        }
    }
    None
}

fn resolve_binary(p: &Path) -> PathBuf {
    if p.exists() {
        return p.to_path_buf();
    }
    find_pryx_core_binary().unwrap_or_else(|| p.to_path_buf())
}

fn restart_backoff(attempt: u32) -> Duration {
    match attempt {
        0 => Duration::from_millis(250),
        1 => Duration::from_millis(500),
        2 => Duration::from_secs(1),
        3 => Duration::from_secs(2),
        _ => Duration::from_secs(5),
    }
}

fn parse_listen_addr(line: &str) -> Option<String> {
    if let Some(rest) = line.strip_prefix("PRYX_CORE_LISTEN_ADDR=") {
        let v = rest.trim();
        if !v.is_empty() {
            return Some(v.to_string());
        }
    }

    let prefix = "Listening on ";
    let idx = line.find(prefix)?;
    let addr = line[idx + prefix.len()..].trim();
    if addr.is_empty() {
        return None;
    }
    Some(addr.to_string())
}

fn parse_permission_request(line: &str) -> Option<RpcRequest> {
    let req: RpcRequest = serde_json::from_str(line).ok()?;
    if req.jsonrpc != "2.0" {
        return None;
    }
    if req.method != "permission.request" {
        return None;
    }
    Some(req)
}
