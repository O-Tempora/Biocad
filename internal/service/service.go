package service

import (
	"context"
	"os"
	"path"
	"time"

	"log/slog"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Service struct {
	Db     *mongo.Database
	Logger *slog.Logger
}

func (s *Service) WriteFilesToDB(files []TargetFile) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var mrsh []interface{}
	for _, v := range files {
		bytes, err := bson.Marshal(v)
		if err != nil {
			return err
		}
		mrsh = append(mrsh, bytes)
	}
	_, err := s.Db.Collection("processed_tsv").InsertMany(ctx, mrsh)
	if err != nil {
		return err
	}
	return nil
}

// func (s *Service) GetProcessedFileNames() ([]string, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// }

func (s *Service) RunDirectiryScanner(fileDir, outputDir string) error {
	if err := os.MkdirAll(fileDir, 0744); err != nil {
		return err
	}
	if err := os.MkdirAll(outputDir, 0744); err != nil {
		return err
	}
	ticker := time.Tick(10 * time.Second)
	for {
		select {
		case <-ticker:
			processedFiles, err := s.processDir(fileDir, outputDir)
			if err != nil {
				s.Logger.ErrorContext(context.Background(), err.Error())
				break
			}
			if err = s.WriteFilesToDB(processedFiles); err != nil {
				s.Logger.ErrorContext(context.Background(), err.Error())
			}
		}
	}
}

func (s *Service) processDir(fileDir, outputDir string) ([]TargetFile, error) {
	var tsvs []TargetFile
	entries, err := os.ReadDir(fileDir)
	if err != nil {
		return nil, err
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
			s.Logger.ErrorContext(context.Background(), err.Error())
			continue
		}
		tf := readTsv(file)
		tf.Name = v.Name()
		tsvs = append(tsvs, tf)
	}
	return tsvs, nil
}
