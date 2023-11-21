package internal

type Config struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	Dir    string `yaml:"dir"`
	DbHost string `yaml:"dbhost"`
	DbPort int    `yaml:"dbport"`
}
