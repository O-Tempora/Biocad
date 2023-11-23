package service

import (
	"fmt"
	"os"
	"path"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	Db     *mongo.Client
	Logger *slog.Logger
}

func (s *Service) RunDirectiryScanner(fileDir, outputDir string) error {
	if err := os.MkdirAll(fileDir, 0744); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0744); err != nil {
		return err
	}
	ticker := time.Tick(30 * time.Second)
	for {
		select {
		case <-ticker:
			s.processDir(fileDir, outputDir)
		}
	}
}

func (s *Service) processDir(fileDir, outputDir string) error {
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return err
	}
	for _, v := range entries {
		if v.IsDir() {
			continue
		}
		filePath := path.Join(fileDir, v.Name())
		if path.Ext(filePath) != ".tsv" {
			continue
		}
		file, err := os.OpenFile(filePath, os.O_RDONLY, 0766)
		if err != nil {
			s.Logger.Error("On opening file:", err.Error())
			continue
		}
		tsv := readTsv(file)
		for _, v := range tsv.Records {
			fmt.Println(v)
		}
	}
	return nil
}
