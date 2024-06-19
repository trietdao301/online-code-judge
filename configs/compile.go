package configs

type Compile struct {
	Image           string   `yaml:"image"`
	CommandTemplate []string `yaml:"command_template"`
	Timeout         string   `yaml:"timeout"`
	CPUQuota        int64    `yaml:"cpu_quota"`
	Memory          string   `yaml:"memory"`
	WorkingDir      string   `yaml:"working_dir"`
	SourceFileName  string   `yaml:"source_file_name"`
	ProgramFileName string   `yaml:"program_file_name"`
}
