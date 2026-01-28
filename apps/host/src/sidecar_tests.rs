#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_config() {
        let config = SidecarConfig::default();
        assert_eq!(config.binary_path, "pryx-core");
        assert_eq!(config.max_restarts, 10);
        assert_eq!(config.initial_backoff_ms, 1000);
        assert_eq!(config.max_backoff_ms, 30000);
        assert_eq!(config.backoff_multiplier, 2.0);
        assert_eq!(config.port_discovery_timeout_secs, 10);
    }

    #[test]
    fn test_sidecar_state_variants() {
        assert!(matches!(SidecarState::Stopped, SidecarState::Stopped));
        assert!(matches!(SidecarState::Starting, SidecarState::Starting));
        assert!(matches!(SidecarState::Running, SidecarState::Running));
        assert!(matches!(SidecarState::Crashed { attempts: 0 }, SidecarState::Crashed { attempts: 0 }));
        assert!(matches!(SidecarState::Restarting { backoff_ms: 0 }, SidecarState::Restarting { backoff_ms: 0 }));
        assert!(matches!(SidecarState::Stopping, SidecarState::Stopping));
    }

    #[test]
    fn test_sidecar_status_serialization() {
        let status = SidecarStatus {
            state: SidecarState::Running,
            pid: Some(12345),
            port: Some(8080),
            uptime_secs: Some(100.5),
            crash_count: 2,
            started_at: Some("1234567890".to_string()),
        };

        let serialized = serde_json::to_string(&status).unwrap();
        let deserialized: SidecarStatus = serde_json::from_str(&serialized).unwrap();

        assert_eq!(deserialized.state, status.state);
        assert_eq!(deserialized.pid, status.pid);
        assert_eq!(deserialized.port, status.port);
        assert_eq!(deserialized.uptime_secs, status.uptime_secs);
    }

    #[test]
    fn test_extract_port_from_listening() {
        assert_eq!(extract_port_from_line("Listening on :8080"), Some(8080));
        assert_eq!(extract_port_from_line("Server started on port 8080"), Some(8080));
        assert_eq!(extract_port_from_line("port:8080"), Some(8080));
        assert_eq!(extract_port_from_line("listening on 0.0.0.0:8080"), Some(8080));
        assert_eq!(extract_port_from_line("listening port: 8080"), Some(8080));
    }

    #[test]
    fn test_extract_port_from_no_match() {
        assert_eq!(extract_port_from_line("Hello world"), None);
        assert_eq!(extract_port_from_line("No port here"), None);
        assert_eq!(extract_port_from_line(""), None);
    }

    #[test]
    fn test_calculate_backoff() {
        let config = SidecarConfig::default();
        
        // First attempt: no backoff
        let backoff1 = calculate_backoff(1, &config);
        assert_eq!(backoff1, config.initial_backoff_ms);
        
        // Second attempt: 2x
        let backoff2 = calculate_backoff(2, &config);
        assert_eq!(backoff2, config.initial_backoff_ms * config.backoff_multiplier);
        
        // Third attempt: 4x
        let backoff3 = calculate_backoff(3, &config);
        assert_eq!(backoff3, config.initial_backoff_ms * config.backoff_multiplier.powi(2));
        
        // Tenth attempt: 2^9 = 512x
        let backoff10 = calculate_backoff(10, &config);
        assert_eq!(backoff10, config.initial_backoff_ms * config.backoff_multiplier.powi(9));
        
        // Should cap at max_backoff
        let backoff20 = calculate_backoff(20, &config);
        let expected = config.initial_backoff_ms * config.backoff_multiplier.powi(10); // Capped
        assert_eq!(backoff20, expected);
    }
}
