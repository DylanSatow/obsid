package config

type Config struct {
	Vault     VaultConfig     `yaml:"vault"`
	Projects  ProjectsConfig  `yaml:"projects"`
	Templates TemplatesConfig `yaml:"templates"`
	Git       GitConfig       `yaml:"git"`
	Format    FormatConfig    `yaml:"formatting"`
}

type VaultConfig struct {
	Path          string `yaml:"path"`
	DailyNotesDir string `yaml:"daily_notes_dir"`
	DateFormat    string `yaml:"date_format"`
}

type ProjectsConfig struct {
	AutoDiscover bool     `yaml:"auto_discover"`
	Directories  []string `yaml:"directories"`
}

type TemplatesConfig struct {
	ProjectEntry string `yaml:"project_entry"`
}

type GitConfig struct {
	IncludeDiffs       bool `yaml:"include_diffs"`
	MaxCommits         int  `yaml:"max_commits"`
	IgnoreMergeCommits bool `yaml:"ignore_merge_commits"`
}

type FormatConfig struct {
	CreateLinks     bool     `yaml:"create_links"`
	AddTags         []string `yaml:"add_tags"`
	TimestampFormat string   `yaml:"timestamp_format"`
}