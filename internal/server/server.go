package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/O-Tempora/Biocad/internal"
	"github.com/O-Tempora/Biocad/internal/service"
	"github.com/go-chi/chi/v5"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type server struct {
	logger  *slog.Logger
	router  *chi.Mux
	service *service.Service
}

func InitServer(cf *internal.Config) (*server, error) {
	s := &server{
		logger: InitLogger(),
	}
	s.InitRouter()
	db, err := InitDatabase(cf)
	if err != nil {
		return nil, err
	}
	s.service = &service.Service{
		Db:     db.Database("Biocad"),
		Logger: s.logger,
	}
	return s, nil
}

func InitLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func InitDatabase(cf *internal.Config) (*mongo.Client, error) {
	connstr := fmt.Sprintf("mongodb://%s:%d", cf.DbHost, cf.DbPort)
	db, err := mongo.Connect(context.Background(), options.Client().ApplyURI(connstr))
	if err != nil {
		return nil, err
	}
	if err = db.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, err
	}
	return db, nil
}

func (s *server) Service() *service.Service {
	return s.service
}
