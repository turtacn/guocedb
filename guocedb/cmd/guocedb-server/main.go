package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/turtacn/guocedb/common/config"
	"github.com/turtacn/guocedb/common/log"
	"github.com/turtacn/guocedb/compute/executor"
	mysql_server "github.com/turtacn/guocedb/protocol/mysql"
	"github.com/turtacn/guocedb/storage/sal"
	_ "github.com/turtacn/guocedb/storage/engines/badger" // Register the badger engine
)

var (
	configFile = flag.String("config", "./configs/config.yaml.example", "Path to the configuration file.")
)

func main() {
	flag.Parse()

	// 1. Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Initialize logger
	log.SetLevel(cfg.Log.Level)
	log.Infof("guocedb version %s starting...", config.GlobalConfig.Server.Version)

	// 3. Initialize storage engine
	storage, err := sal.NewStorageEngine(&cfg.Storage)
	if err != nil {
		log.Fatalf("Failed to initialize storage engine: %v", err)
	}
	defer storage.Close()
	log.Infof("Using storage engine: %s", cfg.Storage.Engine)

	// 4. Initialize compute engine
	computeEngine := executor.NewEngine(storage)

	// 5. Start the MySQL protocol server
	mysqlSrv, err := mysql_server.NewServer("0.0.0.0", cfg.Server.MySQLPort, computeEngine)
	if err != nil {
		log.Fatalf("Failed to start MySQL server: %v", err)
	}

	go func() {
		if err := mysqlSrv.Start(); err != nil {
			log.Errorf("MySQL server error: %v", err)
		}
	}()

	// TODO: Start gRPC and HTTP servers for management APIs.

	// 6. Wait for shutdown signal
	waitForShutdown(mysqlSrv)
}

func waitForShutdown(srv *mysql_server.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan

	log.Infof("Shutdown signal received. Shutting down servers...")
	if err := srv.Stop(); err != nil {
		log.Errorf("Error shutting down MySQL server: %v", err)
	}
	log.Infof("Shutdown complete.")
}
