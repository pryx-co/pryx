package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

const launchdTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.pryx.runtime</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>{{.LogDir}}/runtime.log</string>
    <key>StandardErrorPath</key>
    <string>{{.LogDir}}/runtime.err.log</string>
    <key>EnvironmentVariables</key>
    <dict>
        <key>PATH</key>
        <string>/usr/local/bin:/usr/bin:/bin:/usr/sbin:/sbin</string>
    </dict>
</dict>
</plist>`

const systemdTemplate = `[Unit]
Description=Pryx Runtime Service
After=network.target

[Service]
ExecStart={{.BinaryPath}}
Restart=always
RestartSec=5
User={{.User}}
Environment=HOME={{.Home}}
StandardOutput=append:{{.LogDir}}/runtime.log
StandardError=append:{{.LogDir}}/runtime.err.log

[Install]
WantedBy=multi-user.target`

type ServiceConfig struct {
	BinaryPath string
	LogDir     string
	User       string
	Home       string
}

func runInstallService() int {
	fmt.Println("Installing Pryx system service...")

	binaryPath, err := os.Executable()
	if err != nil {
		log.Printf("Failed to get executable path: %v", err)
		return 1
	}
	// Use absolute path
	binaryPath, _ = filepath.Abs(binaryPath)

	home, _ := os.UserHomeDir()
	logDir := filepath.Join(home, ".pryx", "logs")
	os.MkdirAll(logDir, 0755)

	sc := ServiceConfig{
		BinaryPath: binaryPath,
		LogDir:     logDir,
		User:       os.Getenv("USER"),
		Home:       home,
	}

	switch runtime.GOOS {
	case "darwin":
		return installLaunchd(sc)
	case "linux":
		return installSystemd(sc)
	default:
		fmt.Printf("Install service not supported on %s\n", runtime.GOOS)
		return 1
	}
}

func installLaunchd(sc ServiceConfig) int {
	plistName := "com.pryx.runtime.plist"
	plistPath := filepath.Join(sc.Home, "Library/LaunchAgents", plistName)

	tmpl, _ := template.New("launchd").Parse(launchdTemplate)
	f, err := os.Create(plistPath)
	if err != nil {
		log.Printf("Failed to create plist file: %v", err)
		return 1
	}
	defer f.Close()

	if err := tmpl.Execute(f, sc); err != nil {
		log.Printf("Failed to generate plist: %v", err)
		return 1
	}

	// Load the service
	cmd := exec.Command("launchctl", "load", plistPath)
	if err := cmd.Run(); err != nil {
		log.Printf("Failed to load service with launchctl: %v", err)
		return 1
	}

	fmt.Println("✓ Pryx service installed and started (launchd)")
	fmt.Printf("  Service file: %s\n", plistPath)
	fmt.Printf("  Logs: %s/runtime.log\n", sc.LogDir)
	return 0
}

func installSystemd(sc ServiceConfig) int {
	serviceName := "pryx.service"
	servicePath := "/etc/systemd/system/" + serviceName

	tmpl, _ := template.New("systemd").Parse(systemdTemplate)

	// Write to temp file first then sudo mv
	tmpFile := filepath.Join(os.TempDir(), serviceName)
	f, err := os.Create(tmpFile)
	if err != nil {
		log.Printf("Failed to create temp service file: %v", err)
		return 1
	}

	if err := tmpl.Execute(f, sc); err != nil {
		f.Close()
		log.Printf("Failed to generate service file: %v", err)
		return 1
	}
	f.Close()

	fmt.Println("Service file generated. Requires sudo to install to /etc/systemd/system/")

	// Move to system location
	if err := exec.Command("sudo", "mv", tmpFile, servicePath).Run(); err != nil {
		log.Printf("Failed to move service file (sudo required): %v", err)
		return 1
	}

	// Reload, enable and start
	_ = exec.Command("sudo", "systemctl", "daemon-reload").Run()
	_ = exec.Command("sudo", "systemctl", "enable", serviceName).Run()
	if err := exec.Command("sudo", "systemctl", "start", serviceName).Run(); err != nil {
		log.Printf("Failed to start service: %v", err)
		return 1
	}

	fmt.Println("✓ Pryx service installed and started (systemd)")
	return 0
}
