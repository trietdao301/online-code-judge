package configs

type Logic struct {
	Judge Judge `yaml:"judge"`
}

type TestCaseRun struct {
	Image           string   `yaml:"image"`
	CommandTemplate []string `yaml:"command_template"`
	CPUQuota        int64    `yaml:"cpu_quota"`
	CodeFileName    string   `yaml:"code_file_name"`
	TestFileName    string   `yaml:"test_file_name"`
	StdOut          bool     `yaml:"stdOut"`
	StdErr          bool     `yaml:"stdErr"`
	DownloadTestUrl *string  `yaml:"download_test_url,omitempty"`
	TestLibraryName *string  `yaml:"test_library_name,omitempty"`
}

type Language struct {
	Value       string      `yaml:"value"`
	Name        string      `yaml:"name"`
	TestCaseRun TestCaseRun `yaml:"test_case_run"`
}
type Judge struct {
	Schedule             string     `yaml:"schedule"`
	Languages            []Language `yaml:"languages"`
	SubmissionRetryDelay string     `yaml:"submission_retry_delay"`
}
