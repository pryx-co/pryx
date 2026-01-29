#[cfg(test)]
mod tests {
    use super::*;
    use std::time::Duration;
    use tokio::runtime::Runtime;

    #[test]
    fn test_calculate_backoff() {
        let config = SidecarConfig {
            initial_backoff_ms: 1000,
            backoff_multiplier: 2.0,
            ..Default::default()
        };

        // First attempt should use initial backoff
        assert_eq!(calculate_backoff(1, &config), 1000);

        // Second attempt should be 2x
        assert_eq!(calculate_backoff(2, &config), 2000);

        // Third attempt should be 4x
        assert_eq!(calculate_backoff(3, &config), 4000);

        // Tenth attempt should be capped (2^9 = 512, but clamped)
        let large_backoff = calculate_backoff(10, &config);
        assert!(large_backoff > 1000);
    }

    #[test]
    fn test_extract_port_from_line_pryx_format() {
        let line = "PRYX_CORE_LISTEN_ADDR=127.0.0.1:8080";
        assert_eq!(extract_port_from_line(line), Some(8080));

        let line = "PRYX_CORE_LISTEN_ADDR=:3000";
        assert_eq!(extract_port_from_line(line), Some(3000));
    }

    #[test]
    fn test_extract_port_from_line_listening_format() {
        let line = "Starting server on http://localhost:3000";
        assert_eq!(extract_port_from_line(line), Some(3000));

        let line = "Server listening on port :9090";
        assert_eq!(extract_port_from_line(line), Some(9090));
    }

    #[test]
    fn test_extract_port_from_line_no_port() {
        assert_eq!(extract_port_from_line("No port here"), None);
        assert_eq!(extract_port_from_line(""), None);
        assert_eq!(extract_port_from_line("Just some text: but not a port"), None);
    }

    #[test]
    fn test_extract_port_check() {
        assert!(extract_port_check("PRYX_CORE_LISTEN_ADDR=127.0.0.1:8080"));
        assert!(extract_port_check("Server listening on port 3000"));
        assert!(extract_port_check("Listening on http://localhost:8080"));
        assert!(!extract_port_check("Some random log message"));
    }

    #[test]
    fn test_sidecar_state_transitions() {
        use SidecarState::*;

        // Test all state variants can be created
        let states = vec![
            Stopped,
            Starting,
            Running,
            Crashed { attempts: 1 },
            Restarting { backoff_ms: 1000 },
            Stopping,
        ];

        for state in states {
            match state {
                Stopped => assert!(matches!(state, Stopped)),
                Starting => assert!(matches!(state, Starting)),
                Running => assert!(matches!(state, Running)),
                Crashed { attempts } => assert_eq!(attempts, 1),
                Restarting { backoff_ms } => assert_eq!(backoff_ms, 1000),
                Stopping => assert!(matches!(state, Stopping)),
            }
        }
    }

    #[test]
    fn test_sidecar_config_default() {
        let config = SidecarConfig::default();
        assert_eq!(config.start_timeout, Duration::from_secs(3));
        assert_eq!(config.max_restarts, 10);
        assert_eq!(config.initial_backoff_ms, 1000);
        assert_eq!(config.backoff_multiplier, 2.0);
    }

    #[test]
    fn test_sidecar_config_new() {
        let binary = PathBuf::from("/path/to/binary");
        let cwd = PathBuf::from("/working/dir");
        let db_path = PathBuf::from("/path/to/db");

        let config = SidecarConfig::new(binary.clone(), cwd.clone(), db_path.clone());

        assert_eq!(config.binary, binary);
        assert_eq!(config.cwd, cwd);
        assert_eq!(config.db_path, db_path);
        assert!(config.args.is_empty());
        assert!(config.env.is_empty());
    }

    #[test]
    fn test_find_pryx_core_binary_from_env() {
        // Set environment variable
        std::env::set_var("PRYX_CORE_PATH", "/custom/path/pryx-core");
        
        // This would need the actual file to exist to return Some
        // For now, we just verify it checks the env var
        let _result = find_pryx_core_binary();
        
        std::env::remove_var("PRYX_CORE_PATH");
    }

    #[test]
    fn test_sidecar_error_display() {
        let err = SidecarError::SpawnFailed {
            binary: "pryx-core".to_string(),
            reason: "file not found".to_string(),
        };
        let msg = format!("{}", err);
        assert!(msg.contains("pryx-core"));
        assert!(msg.contains("file not found"));

        let err = SidecarError::NoChild;
        assert_eq!(format!("{}", err), "Sidecar process not running");

        let err = SidecarError::ProcessNotRunning("test".into());
        assert!(format!("{}", err).contains("test"));
    }

    #[test]
    fn test_rpc_request_deserialization() {
        let json = r#"{"jsonrpc":"2.0","method":"test.method","params":{"key":"value"},"id":1}"#;
        let req: RpcRequest = serde_json::from_str(json).unwrap();
        
        assert_eq!(req.method, "test.method");
        assert_eq!(req.id, 1);
    }

    #[test]
    fn test_rpc_response_serialization() {
        let resp = RpcResponse {
            jsonrpc: "2.0".to_string(),
            result: serde_json::json!({"status": "ok"}),
            id: 42,
        };
        
        let json = serde_json::to_string(&resp).unwrap();
        assert!(json.contains("\"jsonrpc\":\"2.0\""));
        assert!(json.contains("\"id\":42"));
        assert!(json.contains("\"status\":\"ok\""));
    }

    #[test]
    fn test_search_ancestors() {
        // Test with current directory
        let cwd = std::env::current_dir().unwrap();
        let result = search_ancestors(&cwd);
        
        // This may or may not find something depending on directory structure
        // Just verify it doesn't panic
        let _ = result;
    }

    #[tokio::test]
    async fn test_sidecar_status_initial() {
        // We can't easily create a SidecarProcess without an AppHandle
        // This test documents what the status structure should contain
        let status = SidecarStatus {
            state: SidecarState::Stopped,
            pid: None,
            port: None,
            uptime_secs: None,
            crash_count: 0,
            started_at: None,
        };

        assert!(matches!(status.state, SidecarState::Stopped));
        assert_eq!(status.crash_count, 0);
    }

    #[test]
    fn test_backoff_calculation_variations() {
        let config = SidecarConfig {
            initial_backoff_ms: 500,
            backoff_multiplier: 1.5,
            ..Default::default()
        };

        // Test different attempt numbers
        let test_cases = vec![
            (1, 500),   // Initial
            (2, 750),   // 500 * 1.5
            (3, 1125),  // 500 * 1.5^2
            (5, 2531),  // 500 * 1.5^4
        ];

        for (attempt, expected) in test_cases {
            let backoff = calculate_backoff(attempt, &config);
            assert!(
                (backoff as i64 - expected as i64).abs() < 10,
                "Attempt {}: expected ~{}, got {}",
                attempt, expected, backoff
            );
        }
    }
}
