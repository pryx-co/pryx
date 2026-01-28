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

type Installer struct {
	ID      string   `yaml:"id" json:"id"`
	Kind    string   `yaml:"kind" json:"kind"`
	Formula string   `yaml:"formula,omitempty" json:"formula,omitempty"`
	Bins    []string `yaml:"bins,omitempty" json:"bins,omitempty"`
	URL     string   `yaml:"url,omitempty" json:"url,omitempty"`
	Args    []string `yaml:"args,omitempty" json:"args,omitempty"`
}

type PryxMetadata struct {
	Emoji    string       `yaml:"emoji,omitempty" json:"emoji,omitempty"`
	Requires Requirements `yaml:"requires,omitempty" json:"requires,omitempty"`
	Install  []Installer  `yaml:"install,omitempty" json:"install,omitempty"`
}

type Metadata struct {
	Pryx PryxMetadata `yaml:"pryx,omitempty" json:"pryx,omitempty"`
}

type Frontmatter struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Metadata    Metadata `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type Skill struct {
	ID          string      `json:"id"`
	Source      Source      `json:"source"`
	Path        string      `json:"path"`
	Frontmatter Frontmatter `json:"frontmatter"`

	bodyLoader func() (string, error)
}

func (s Skill) Body() (string, error) {
	if s.bodyLoader == nil {
		return "", nil
	}
	return s.bodyLoader()
}
