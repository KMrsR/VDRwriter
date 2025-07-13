package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/tbrandon/mbserver"
)

func main() {
	var regMutex sync.Mutex  // определяем мьютекс защиты регистров
	var connMutex sync.Mutex // определяем мьютекс защиты соеднинения с терминальным сервером
	oldTable := make([]uint16, 1500)
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
	fmt.Println("opening ModbusTCP server on localhost:502")
	serv := mbserver.NewServer()
	err = serv.ListenTCP(":502")
	if err != nil {
		log.Printf("%v\nPress 'q' to quit\n", err)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			exit := scanner.Text()
			if exit == "q" {
				os.Exit(1)
			} else {
				fmt.Println("Press 'q' to quit")
			}
		}
	}
	log.Println("ModbusTCP server opened")
	for i := 1; i < 1500; i++ {
		oldTable[i] = serv.HoldingRegisters[i]
	}
	defer serv.Close()
	//--------------------------------------------------------------------
	//opening TCP client
	log.Println("connecting to Terminal server on 10.1.2.65:8007")
	conn, err := net.Dial("tcp", "10.1.2.65:8007")

	if err != nil {
		log.Printf("%v\nPress 'q' to quit\n", err)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			exit := scanner.Text()
			if exit == "q" {
				os.Exit(1)
			} else {
				fmt.Println("Press 'q' to quit")
			}
		}
	}
	log.Println("connection to Terminal server opened")
	defer conn.Close()
	//--------------------------------------------------------------------
	// go iaswd(conn, &connMutex, cfg)
	// go MonitorTags(conn, serv, &regMutex, &connMutex, tagCfg, cfg)
	go WriteAll(conn, serv, &regMutex, &connMutex, tagCfg, cfg)

	select {} // блокировка main
	//--------------------------------------------------------------------
}
