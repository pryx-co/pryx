package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
)

func runUninstallService() int {
	fmt.Println("Attempting to uninstall Pryx system service...")

	switch runtime.GOOS {
	case "darwin":
		return uninstallLaunchd()
	case "linux":
		return uninstallSystemd()
	default:
		fmt.Printf("Uninstall service not supported on %s\n", runtime.GOOS)
		return 1
	}
}

func uninstallLaunchd() int {
	plistName := "com.pryx.runtime.plist"
	home, _ := os.UserHomeDir()
	plistPath := fmt.Sprintf("%s/Library/LaunchAgents/%s", home, plistName)

	// Unload the service
	cmd := exec.Command("launchctl", "unload", plistPath)
	_ = cmd.Run() // Ignore error if already unloaded

	// Remove the plist file
	if err := os.Remove(plistPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("✓ Service file not found (already uninstalled)")
			return 0
		}
		log.Printf("Failed to remove service file: %v", err)
		return 1
	}

	fmt.Println("✓ Pryx service uninstalled successfully (launchd)")
	return 0
}

func uninstallSystemd() int {
	serviceName := "pryx.service"
	servicePath := fmt.Sprintf("/etc/systemd/system/%s", serviceName)

	// Stop the service
	cmd := exec.Command("sudo", "systemctl", "stop", serviceName)
	_ = cmd.Run()

	// Disable the service
	cmd = exec.Command("sudo", "systemctl", "disable", serviceName)
	_ = cmd.Run()

	// Remove the service file
	if err := exec.Command("sudo", "rm", servicePath).Run(); err != nil {
		log.Printf("Failed to remove service file: %v", err)
		return 1
	}

	// Reload systemd
	_ = exec.Command("sudo", "systemctl", "daemon-reload").Run()

	fmt.Println("✓ Pryx service uninstalled successfully (systemd)")
	return 0
}
