package skills

type Source string

const (
	SourceBundled   Source = "bundled"
	SourceManaged   Source = "managed"
	SourceWorkspace Source = "workspace"
)

type Requirements struct {
	Bins []string `yaml:"bins" json:"bins"`
	Env  []string `yaml:"env" json:"env"`
}

type SkillMetadata struct {
	Pryx PryxMetadata `yaml:"pryx,omitempty" json:"pryx,omitempty"`
}

type Frontmatter struct {
	Name        string        `yaml:"name"`
	Description string        `yaml:"description"`
	Metadata    SkillMetadata `yaml:"metadata,omitempty"`
}

type Installer struct {
	ID        string   `yaml:"id" json:"id"`
	Kind      string   `yaml:"kind" json:"kind"`
	Command   string   `yaml:"command" json:"command"`
	Args      []string `yaml:"args,omitempty" json:"args,omitempty"`
	URL       string   `yaml:"url,omitempty" json:"url,omitempty"`
	Directory string   `yaml:"directory,omitempty" json:"directory,omitempty"`
	Env       []string `yaml:"env,omitempty" json:"env,omitempty"`
}

type PryxMetadata struct {
	Emoji    string       `yaml:"emoji,omitempty" json:"emoji,omitempty"`
	Requires Requirements `yaml:"requires,omitempty" json:"requires,omitempty"`
	Install  []Installer  `yaml:"install,omitempty" json:"install,omitempty"`
}

type Skill struct {
	ID           string                 `yaml:"id" json:"id"`
	Source       Source                 `yaml:"source" json:"source"`
	Name         string                 `yaml:"name" json:"name"`
	Title        string                 `yaml:"title,omitempty" json:"title"`
	Description  string                 `yaml:"description" json:"description"`
	Version      string                 `yaml:"version" json:"version"`
	Author       string                 `yaml:"author,omitempty" json:"author"`
	Path         string                 `yaml:"path" json:"path"`
	Frontmatter  Frontmatter            `yaml:"frontmatter"`
	Enabled      bool                   `yaml:"enabled" json:"enabled"`
	Eligible     bool                   `yaml:"eligible" json:"eligible"`
	SystemPrompt string                 `yaml:"system_prompt,omitempty" json:"system_prompt"`
	UserPrompt   string                 `yaml:"user_prompt,omitempty" json:"user_prompt"`
	Metadata     map[string]interface{} `yaml:"metadata,omitempty" json:"metadata"`

	bodyLoader func() (string, error)
}

func (s Skill) Parse() (Frontmatter, error) {
	return Frontmatter{}, nil
}

func (s Skill) Body() (string, error) {
	if s.bodyLoader != nil {
		return s.bodyLoader()
	}
	return "", nil
}
