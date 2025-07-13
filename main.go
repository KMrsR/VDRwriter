package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/tbrandon/mbserver"
)

func main() {
	var regMutex sync.Mutex  // определяем мьютекс защиты регистров
	var connMutex sync.Mutex // определяем мьютекс защиты соеднинения с терминальным сервером
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	//--------------------------------------------------------------------
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received OS signal: %s", sig)
		cancel()
	}()
	//--------------------------------------------------------------------
	// читаем конфиги
	cfg, err := loadConfig("./config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	tagCfg, err := loadTagConfig("./map.yaml")
	if err != nil {
		log.Fatalf("Failed to load map config: %v", err)
	}
	//--------------------------------------------------------------------
	//Modbus TCP server
	serv := mbserver.NewServer()
	err = serv.ListenTCP(cfg.MBTCPip)
	if err != nil {
		log.Printf("Failed to open ModbusTCP server on %s: %v", cfg.MBTCPip, err)
		waitForExitOnWindows()
		os.Exit(1)
	}
	log.Printf("ModbusTCP server on %s started", cfg.MBTCPip)
	//--------------------------------------------------------------------
	//opening TCP client
	conn, err := net.Dial("tcp", cfg.ETOSip)
	if err != nil {
		log.Printf("Failed to connect to Terminal server %s: %v", cfg.ETOSip, err)
		waitForExitOnWindows()
		os.Exit(1)
	}
	log.Printf("connected to Terminal server on %s", cfg.ETOSip)
	//--------------------------------------------------------------------
	go MonitorTags(ctx, conn, serv, &regMutex, &connMutex, tagCfg, cfg)
	go WriteAll(ctx, conn, serv, &regMutex, &connMutex, tagCfg, cfg)
	go iaswd(ctx, conn, serv, &regMutex, &connMutex, tagCfg, cfg)
	//--------------------------------------------------------------------
	<-ctx.Done()
	defer conn.Close()
	defer serv.Close()
	log.Println("Shutdown signal received, closing resources...")
	//--------------------------------------------------------------------
	waitForExitOnWindows()
}
