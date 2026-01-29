package skills

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type Options struct {
	WorkspaceRoot string
	ManagedRoot   string
	BundledRoot   string
	MaxConcurrent int
}

func DefaultOptions() Options {
	wd, _ := os.Getwd()
	home, _ := os.UserHomeDir()
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)

	workspaceRoot := strings.TrimSpace(os.Getenv("PRYX_WORKSPACE_ROOT"))
	if workspaceRoot == "" {
		workspaceRoot = wd
	}
	managedRoot := strings.TrimSpace(os.Getenv("PRYX_MANAGED_SKILLS_DIR"))
	if managedRoot == "" {
		managedRoot = filepath.Join(home, ".pryx", "skills")
	}
	bundledRoot := strings.TrimSpace(os.Getenv("PRYX_BUNDLED_SKILLS_DIR"))
	if bundledRoot == "" {
		bundledRoot = findBundledSkillsDir(workspaceRoot)
		if bundledRoot == "" {
			bundledRoot = filepath.Join(execDir, "bundled-skills")
		}
	}
	return Options{
		WorkspaceRoot: workspaceRoot,
		ManagedRoot:   managedRoot,
		BundledRoot:   bundledRoot,
		MaxConcurrent: runtime.GOMAXPROCS(0),
	}
}

func findBundledSkillsDir(workspaceRoot string) string {
	dir := workspaceRoot
	for i := 0; i < 6; i++ {
		candidates := []string{
			filepath.Join(dir, "apps", "runtime", "internal", "skills", "bundled"),
			filepath.Join(dir, "runtime", "internal", "skills", "bundled"),
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				return candidate
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

type MultiError struct {
	Errors []error
}

func (m MultiError) Error() string {
	if len(m.Errors) == 0 {
		return ""
	}
	var b strings.Builder
	for i, err := range m.Errors {
		if err == nil {
			continue
		}
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(err.Error())
	}
	return b.String()
}

func (m MultiError) Unwrap() error {
	if len(m.Errors) == 0 {
		return nil
	}
	return m.Errors[0]
}

func Discover(ctx context.Context, opts Options) (*Registry, error) {
	if opts.MaxConcurrent <= 0 {
		opts.MaxConcurrent = 1
	}

	r := NewRegistry()
	var errs []error

	sources := []struct {
		source Source
		root   string
	}{
		{source: SourceBundled, root: opts.BundledRoot},
		{source: SourceManaged, root: opts.ManagedRoot},
		{source: SourceWorkspace, root: filepath.Join(opts.WorkspaceRoot, ".pryx", "skills")},
	}

	for _, src := range sources {
		if strings.TrimSpace(src.root) == "" {
			continue
		}
		paths, err := findSkillFiles(ctx, src.root)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		m, loadErrs := loadSkillsFromPaths(ctx, src.source, paths, opts.MaxConcurrent)
		errs = append(errs, loadErrs...)
		for _, skill := range m {
			r.Upsert(skill)
		}
	}

	if len(errs) > 0 {
		return r, MultiError{Errors: errs}
	}
	return r, nil
}

func findSkillFiles(ctx context.Context, root string) ([]string, error) {
	info, err := os.Stat(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if !info.IsDir() {
		return nil, nil
	}

	var paths []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if d.IsDir() {
			return nil
		}
		if strings.EqualFold(d.Name(), "SKILL.md") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, err
		}
		return nil, err
	}
	return paths, nil
}

func loadSkillsFromPaths(ctx context.Context, source Source, paths []string, maxConcurrent int) (map[string]Skill, []error) {
	out := map[string]Skill{}
	var errsMu sync.Mutex
	var errs []error
	var outMu sync.Mutex

	sem := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup

	for _, path := range paths {
		if ctx.Err() != nil {
			break
		}
		path := path
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			skill, err := loadSkillFromFile(source, path)
			if err != nil {
				errsMu.Lock()
				errs = append(errs, err)
				errsMu.Unlock()
				return
			}
			outMu.Lock()
			out[skill.ID] = skill
			outMu.Unlock()
		}()
	}

	wg.Wait()
	return out, errs
}

func loadSkillFromFile(source Source, path string) (Skill, error) {
	f, err := os.Open(path)
	if err != nil {
		return Skill{}, err
	}
	defer f.Close()

	fm, err := parseFrontmatterOnly(f)
	if err != nil {
		return Skill{}, err
	}
	id := fm.Name

	return Skill{
		ID:          id,
		Source:      source,
		Path:        path,
		Frontmatter: fm,
		bodyLoader: func() (string, error) {
			b, err := os.ReadFile(path)
			if err != nil {
				return "", err
			}
			_, body, err := parseSkillFile(b)
			if err != nil {
				return "", err
			}
			return body, nil
		},
	}, nil
}
