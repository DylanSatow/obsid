package config

type Config struct {
	Vault      VaultConfig     `yaml:"vault" mapstructure:"vault"`
	Projects   ProjectsConfig  `yaml:"projects" mapstructure:"projects"`
	Templates  TemplatesConfig `yaml:"templates" mapstructure:"templates"`
	Git        GitConfig       `yaml:"git" mapstructure:"git"`
	Formatting FormatConfig    `yaml:"formatting" mapstructure:"formatting"`
}

type VaultConfig struct {
	Path          string `yaml:"path" mapstructure:"path"`
	DailyNotesDir string `yaml:"daily_notes_dir" mapstructure:"daily_notes_dir"`
	DateFormat    string `yaml:"date_format" mapstructure:"date_format"`
}

type ProjectsConfig struct {
	AutoDiscover bool     `yaml:"auto_discover" mapstructure:"auto_discover"`
	Directories  []string `yaml:"directories" mapstructure:"directories"`
}

type TemplatesConfig struct {
	ProjectEntry string `yaml:"project_entry" mapstructure:"project_entry"`
}

type GitConfig struct {
	IncludeDiffs       bool `yaml:"include_diffs" mapstructure:"include_diffs"`
	MaxCommits         int  `yaml:"max_commits" mapstructure:"max_commits"`
	IgnoreMergeCommits bool `yaml:"ignore_merge_commits" mapstructure:"ignore_merge_commits"`
}

type FormatConfig struct {
	CreateLinks     bool     `yaml:"create_links" mapstructure:"create_links"`
	AddTags         []string `yaml:"add_tags" mapstructure:"add_tags"`
	TimestampFormat string   `yaml:"timestamp_format" mapstructure:"timestamp_format"`
}