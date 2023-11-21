package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/O-Tempora/Biocad/internal"
	"github.com/O-Tempora/Biocad/internal/server"
	"gopkg.in/yaml.v3"
)

const configPath = "config/default.yaml"

var (
	dir       string
	outputDir string
	host      string
	port      int
	dbhost    string
	dbport    int
)

func init() {
	flag.StringVar(&dir, "dir", "files", "File directory")
	flag.StringVar(&outputDir, "odir", "processed", "Output file directory")
	flag.StringVar(&host, "host", "localhost", "Application host")
	flag.IntVar(&port, "port", 7999, "Application port")
	flag.StringVar(&dbhost, "dbhost", "localhost", "Database host")
	flag.IntVar(&dbport, "dbport", 7998, "Database port")
}
func main() {
	flag.Parse()
	cf := &internal.Config{}
	file, err := os.OpenFile(configPath, os.O_RDONLY, 0644)
	if err != nil {
		log.Fatal(err.Error())
	}
	dec := yaml.NewDecoder(file)
	if err = dec.Decode(&cf); err != nil {
		log.Fatal(err.Error())
	}
	file.Close()

	cf.SourceDir = dir
	cf.OutputDir = outputDir
	cf.DbHost = dbhost
	cf.DbPort = dbport
	cf.Host = host
	cf.Port = port

	s, err := server.InitServer(cf)
	if err != nil {
		log.Fatal(err.Error())
	}

	go s.Service().RunDirectiryScanner(cf.SourceDir, cf.OutputDir)

	if err = http.ListenAndServe(fmt.Sprintf("%s:%d", cf.Host, cf.Port), s); err != nil {
		log.Fatal(err.Error())
	}
}
