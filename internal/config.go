package internal

type Config struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	SourceDir string `yaml:"sourceDir"`
	OutputDir string `yaml:"outputDir"`
	DbHost    string `yaml:"dbhost"`
	DbPort    int    `yaml:"dbport"`
}
