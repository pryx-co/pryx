package agentbus

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"pryx-core/internal/bus"
)

// PackageManager manages agent package installation and management
type PackageManager struct {
	mu         sync.RWMutex
	bus        *bus.Bus
	logger     *StructuredLogger
	packageDir string
	packages   map[string]*AgentPackage
	running    bool
	stopCh     chan struct{}
}

// NewPackageManager creates a new package manager
func NewPackageManager(b *bus.Bus, packageDir string) *PackageManager {
	if packageDir == "" {
		packageDir = filepath.Join(os.Getenv("HOME"), ".pryx", "packages")
	}

	// Ensure package directory exists
	os.MkdirAll(packageDir, 0755)

	return &PackageManager{
		bus:        b,
		logger:     NewStructuredLogger("packages", "info"),
		packageDir: packageDir,
		packages:   make(map[string]*AgentPackage),
		stopCh:     make(chan struct{}),
	}
}

// Start initializes the package manager
func (pm *PackageManager) Start(ctx context.Context) error {
	pm.mu.Lock()
	if pm.running {
		pm.mu.Unlock()
		return nil
	}
	pm.running = true
	pm.mu.Unlock()

	// Load existing packages
	if err := pm.loadPackages(); err != nil {
		pm.logger.Error("failed to load packages", map[string]interface{}{
			"error": err.Error(),
		})
	}

	pm.logger.Info("package manager started", map[string]interface{}{
		"package_dir": pm.packageDir,
	})
	pm.bus.Publish(bus.NewEvent("agentbus.packages.started", "", nil))

	return nil
}

// Stop gracefully shuts down the package manager
func (pm *PackageManager) Stop(ctx context.Context) error {
	pm.mu.Lock()
	if !pm.running {
		pm.mu.Unlock()
		return nil
	}
	pm.running = false
	pm.mu.Unlock()

	close(pm.stopCh)

	pm.logger.Info("package manager stopped", nil)
	pm.bus.Publish(bus.NewEvent("agentbus.packages.stopped", "", nil))

	return nil
}

// Install installs an agent package
func (pm *PackageManager) Install(ctx context.Context, pkg AgentPackage) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if already installed
	if _, exists := pm.packages[pkg.Name]; exists {
		pm.logger.Warn("package already installed", map[string]interface{}{
			"name": pkg.Name,
		})
		return nil
	}

	// Validate package
	if pkg.Name == "" {
		return nil
	}

	// Create package directory
	pkgDir := filepath.Join(pm.packageDir, pkg.Name)
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		return err
	}

	// Save package metadata
	metadataPath := filepath.Join(pkgDir, "package.json")
	// In a real implementation, would save JSON here
	_ = metadataPath

	// Store package
	pm.packages[pkg.Name] = &pkg

	pm.logger.Info("package installed", map[string]interface{}{
		"name":    pkg.Name,
		"version": pkg.Version,
	})

	pm.bus.Publish(bus.NewEvent("agentbus.package.installed", "", map[string]interface{}{
		"name":    pkg.Name,
		"version": pkg.Version,
	}))

	return nil
}

// Uninstall removes an agent package
func (pm *PackageManager) Uninstall(ctx context.Context, pkg AgentPackage) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if package exists
	if _, exists := pm.packages[pkg.Name]; !exists {
		pm.logger.Warn("package not found", map[string]interface{}{
			"name": pkg.Name,
		})
		return nil
	}

	// Remove package directory
	pkgDir := filepath.Join(pm.packageDir, pkg.Name)
	if err := os.RemoveAll(pkgDir); err != nil {
		return err
	}

	// Remove from registry
	delete(pm.packages, pkg.Name)

	pm.logger.Info("package uninstalled", map[string]interface{}{
		"name": pkg.Name,
	})

	pm.bus.Publish(bus.NewEvent("agentbus.package.uninstalled", "", map[string]interface{}{
		"name": pkg.Name,
	}))

	return nil
}

// List returns all installed packages
func (pm *PackageManager) List(ctx context.Context) []*AgentPackage {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	packages := make([]*AgentPackage, 0, len(pm.packages))
	for _, pkg := range pm.packages {
		packages = append(packages, pkg)
	}

	return packages
}

// Get retrieves a package by name
func (pm *PackageManager) Get(ctx context.Context, name string) (*AgentPackage, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pkg, exists := pm.packages[name]
	if !exists {
		return nil, nil
	}

	return pkg, nil
}

// Search searches for packages
func (pm *PackageManager) Search(ctx context.Context, query string) []*AgentPackage {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	var results []*AgentPackage
	for _, pkg := range pm.packages {
		if containsString(pkg.Name, query) ||
			containsString(pkg.Description, query) {
			results = append(results, pkg)
		}
	}

	return results
}

// Update updates a package
func (pm *PackageManager) Update(ctx context.Context, pkg AgentPackage) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.packages[pkg.Name]; !exists {
		return nil
	}

	pm.packages[pkg.Name] = &pkg

	pm.logger.Info("package updated", map[string]interface{}{
		"name":    pkg.Name,
		"version": pkg.Version,
	})

	return nil
}

// GetPackageDir returns the package directory
func (pm *PackageManager) GetPackageDir() string {
	return pm.packageDir
}

// Count returns the number of installed packages
func (pm *PackageManager) Count() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.packages)
}

// loadPackages loads installed packages from disk
func (pm *PackageManager) loadPackages() error {
	entries, err := os.ReadDir(pm.packageDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// In a real implementation, would load package.json
		pkgName := entry.Name()
		pm.packages[pkgName] = &AgentPackage{
			Name:    pkgName,
			Version: "1.0.0",
		}
	}

	return nil
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsSubstring(s, substr))))
}

// containsSubstring is a helper for substring search
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
