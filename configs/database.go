package configs

type Database struct {
	FilePath        string          `yaml:"file_path"`
	MongoCollection MongoCollection `yaml:"mongo_collection"`
}

type MongoCollection struct {
	Submission string `yaml:"submission"`
	TestCase   string `yaml:"test_case"`
	Problem    string `yaml:"problem"`
	Account    string `yaml:"account"`
}
