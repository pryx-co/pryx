//! Unit tests for pryx-host

use std::path::PathBuf;
use crate::sidecar::{SidecarConfig, SidecarError};

#[test]
fn test_sidecar_config_default() {
    let config = SidecarConfig::default();
    assert_eq!(config.binary, PathBuf::from("pryx-core"));
    assert!(config.args.is_empty());
}

#[test]
fn test_sidecar_config_new() {
    let config = SidecarConfig::new(
        PathBuf::from("/usr/bin/pryx-core"),
        PathBuf::from("/tmp"),
        PathBuf::from("/tmp/pryx.db"),
    );
    assert_eq!(config.binary, PathBuf::from("/usr/bin/pryx-core"));
    assert_eq!(config.cwd, PathBuf::from("/tmp"));
    assert!(config.args.is_empty());
}

#[test]
fn test_sidecar_error_display() {
    let err = SidecarError::ProcessNotRunning("test".into());
    let msg = format!("{}", err);
    assert!(msg.contains("test"));
}

#[test]
fn test_sidecar_error_io() {
    let io_err = std::io::Error::new(std::io::ErrorKind::NotFound, "not found");
    let err = SidecarError::Io(io_err);
    let msg = format!("{:?}", err);
    assert!(msg.contains("Io"));
}
