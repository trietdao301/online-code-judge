package configs

type Database struct {
	Name            string          `yaml:"name"`
	MongoCollection MongoCollection `yaml:"mongo_collection"`
}

type MongoCollection struct {
	Submission        string `yaml:"submission"`
	TestCase          string `yaml:"test_case"`
	Problem           string `yaml:"problem"`
	Account           string `yaml:"account"`
	SubmissionSnippet string `yaml:"submission_snippet"`
}
