package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"database/sql"

	pb "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/api/gen"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/app"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/config"
	"github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/server/grpc"
	memorystorage "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/IvanAndreevichPle/hw12_13_14_15_16_calendar/internal/storage/sql"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/grpc"
)

var configFile string
var migrationsPath string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.yaml", "Path to configuration file")
	flag.StringVar(&migrationsPath, "migrations", "migrations", "Path to migrations directory")
}

func main() {
	flag.Parse()

	configData, err := config.NewConfigFromFile(configFile)
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	logg := logger.New(configData.Logger.Level)

	var storage app.Storage
	switch configData.Storage.Type {
	case "memory":
		storage = memorystorage.New()
	case "sql":
		dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			configData.DB.Host, configData.DB.Port, configData.DB.User, configData.DB.Password, configData.DB.DBName)
		if err := runMigrations(dsn); err != nil {
			panic("failed to apply migrations: " + err.Error())
		}
		storage, err = sqlstorage.New(dsn)
		if err != nil {
			panic("failed to connect to db: " + err.Error())
		}
	default:
		panic("unknown storage type: " + configData.Storage.Type)
	}

	calendar := app.New(logg, storage)

	// Порт для grpc: из переменной окружения или по умолчанию
	grpcPort := os.Getenv("GRPC_PORT")
	if grpcPort == "" {
		grpcPort = ":50051"
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterEventServiceServer(grpcSrv, grpcserver.NewServer(calendar))

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	logg.Info("gRPC server listening on " + grpcPort)

	// Graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer stop()

	go func() {
		<-ctx.Done()
		logg.Info("Shutting down gRPC server...")
		stopped := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
			logg.Info("gRPC server stopped gracefully")
		case <-time.After(3 * time.Second):
			logg.Warn("gRPC server stop timeout, forcing stop")
			grpcSrv.Stop()
		}
	}()

	if err := grpcSrv.Serve(lis); err != nil {
		logg.Error("failed to serve: " + err.Error())
		os.Exit(1)
	}
}

func runMigrations(dsn string) error {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()
	return goose.Up(db, migrationsPath)
}
