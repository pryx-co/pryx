use tauri::AppHandle;
use std::sync::{Arc, Mutex};
use tokio::sync::oneshot::Sender;
use serde::{Deserialize, Serialize};

/// Approval request from UI
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ApprovalRequest {
	pub request_id: String,
	pub tool_name: String,
	pub tool_description: String,
	pub args: serde_json::Value,
	pub session_id: String,
	pub is_critical: bool,
}

/// User response to approval request
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum ApprovalResponse {
	#[serde(rename = "approved")]
	Approved,
	#[serde(rename = "denied")]
	Denied,
	#[serde(rename = "cancelled")]
	Cancelled,
}

/// Permission dialog configuration
#[derive(Debug, Clone)]
pub struct PermissionDialogConfig {
	pub show_native_dialog: bool,
	pub dialog_timeout_ms: u64,
	pub approval_required_for_critical: bool,
}

impl Default for PermissionDialogConfig {
	fn default() -> Self {
		Self {
			show_native_dialog: true,
			dialog_timeout_ms: 500,
			approval_required_for_critical: true,
		}
	}
}

/// Permission manager for handling approvals
pub struct PermissionManager {
	config: PermissionDialogConfig,
	pending_requests: Arc<Mutex<Vec<ApprovalRequest>>>,
	request_senders: Arc<Mutex<std::collections::HashMap<String, Sender<ApprovalResponse>>>>,
}

impl PermissionManager {
	pub fn new(config: PermissionDialogConfig) -> Self {
		Self {
			config,
			pending_requests: Arc::new(Mutex::new(Vec::new())),
			request_senders: Arc::new(Mutex::new(std::collections::HashMap::new())),
		}
	}

	pub async fn request_approval(
		&self,
		app_handle: &AppHandle,
		request: ApprovalRequest,
	) -> Result<ApprovalResponse, String> {
		let request_id = request.request_id.clone();

		// Register response handler
		let (sender, mut receiver) = tokio::sync::oneshot::channel();
		{
			let mut senders = self.request_senders.lock().unwrap();
			senders.insert(request_id.clone(), sender);
		}

		// Check if critical tools always require approval
		let should_show_dialog = self.config.approval_required_for_critical && request.is_critical
			|| self.config.show_native_dialog;

		if should_show_dialog {
			// Show native Tauri dialog
			let message = format!(
				"Tool Request: {}\n\nDescription: {}\n\nCritical: {}\n\nDo you want to proceed?",
				request.tool_name,
				request.tool_description,
				if request.is_critical { "Yes - Critical Operation" } else { "No - Standard Operation" }
			);

			// Use Tauri's native dialog with timeout
			let result = tokio::time::timeout(
				tokio::time::Duration::from_millis(self.config.dialog_timeout_ms),
				async {
					app_handle.dialog()
						.message(message)
						.title("Permission Required")
						.kind(tauri::MessageKind::Warning)
						.ok_button_label("Approve")
						.cancel_button_label("Deny")
				}
				.await
			)
			.await;

			match result {
				Some(true) => Ok(ApprovalResponse::Approved),
				Some(false) => Ok(ApprovalResponse::Denied),
				None => {
					// Dialog was dismissed or timed out
					// Default to denied for safety
					Ok(ApprovalResponse::Denied)
				}
			}
		} else {
			// Auto-approve for non-critical tools
			if request.is_critical {
				return Err("Critical tool requires approval but native dialogs are disabled".to_string());
			} else {
				Ok(ApprovalResponse::Approved)
			}
		}
	}

	pub async fn wait_for_response(
		&self,
		request_id: String,
		timeout_ms: Option<u64>,
	) -> Result<ApprovalResponse, String> {
		let timeout_duration = timeout_ms
			.map_or(30_000) // Default 30s
			.map(|ms| tokio::time::Duration::from_millis(*ms));

		let mut senders = self.request_senders.lock().unwrap();
		let sender = senders.get(&request_id)
			.ok_or_else(|| return Err(format!("No sender registered for request {}", request_id)))?;

		let result = tokio::time::timeout(timeout_duration, sender.recv()).await;

		match result {
			Some(response) => Ok(response),
			None => {
				// Timeout - deny by default
				Err(format!("Approval request {} timed out after {:?}", request_id, timeout_duration))
			}
		}
	}

	pub fn cancel_request(&self, request_id: String) {
		let mut senders = self.request_senders.lock().unwrap();
		let mut pending = self.pending_requests.lock().unwrap();

		// Remove from pending requests
		pending.retain(|r| r.request_id != request_id);

		// Close sender to unblock any waiting receivers
		if let Some(sender) = senders.remove(&request_id) {
			drop(sender);
		}
	}

	pub fn list_pending(&self) -> Vec<ApprovalRequest> {
		let pending = self.pending_requests.lock().unwrap();
		pending.clone()
	}

	pub fn get_config(&self) -> PermissionDialogConfig {
		self.config.clone()
	}

/// Tauri command to request approval
#[tauri::command]
pub async fn request_tool_approval(
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

	// Run approval request in background task
	let result = tokio::task::spawn_blocking(move || {
		manager_clone.request_approval(&app, request_clone)
	}).await?;

	match result {
		Ok(ApprovalResponse::Approved) => Ok(true),
		Ok(ApprovalResponse::Denied) | Ok(ApprovalResponse::Cancelled) => Ok(false),
		Err(e) => Err(e),
	}
}

/// Tauri command to configure permissions
#[tauri::command]
pub async fn configure_permissions(
	app: AppHandle,
	config: PermissionDialogConfig,
) -> Result<(), String> {
	let config_path = app.path().join("permissions.json");

	let config_json = serde_json::to_string_pretty(&config)
		.map_err(|e| e.to_string())?;

	std::fs::write(&config_path, config_json)
		.map_err(|e| e.to_string())?;

	eprintln!("Permissions config updated");
	Ok(())
}

/// Tauri command to list pending approvals
#[tauri::command]
pub async fn list_pending_approvals(app: AppHandle) -> Vec<ApprovalRequest> {
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

	let manager = Arc::new(PermissionManager::new(config));
	manager.list_pending()
}

/// Tauri command to cancel an approval request
#[tauri::command]
pub async fn cancel_approval(app: AppHandle, request_id: String) -> Result<(), String> {
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

	let manager = Arc::new(PermissionManager::new(config));
	manager.cancel_request(request_id);

	eprintln!("Cancelled approval request: {}", request_id);
	Ok(())
}

/// Tauri command to check if native dialogs are supported on this platform
#[tauri::command]
pub async fn check_native_dialog_support(app: AppHandle) -> bool {
	// macOS and Windows have native dialogs
	// Linux may have limited native dialog support depending on desktop environment
	#[cfg(target_os = "macos")]
	return true;

	#[cfg(target_os = "windows")]
	return true;

	#[cfg(target_os = "linux")]
	return false;

/// Tauri command to show a permission dialog
#[tauri::command]
pub async fn show_permission_dialog(app: AppHandle, message: String, title: String) -> Result<bool, String> {
	// Show a native Tauri dialog
	let result = app.dialog()
		.message(message)
		.title(title)
		.kind(tauri::MessageKind::Warning)
		.ok_button_label("Approve")
		.cancel_button_label("Deny")
	.await;

	match result {
		Some(true) => Ok(true),
		Some(false) => Ok(false),
		None => Err("Dialog was dismissed".to_string()),
	}
}
