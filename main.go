package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tbrandon/mbserver"
)

var cnfg config

func main() {

	var mutex sync.Mutex // определяем мьютекс
	oldTable := make([]uint16, 1500)
	//--------------------------------------------------------------------
	//read config
	file, _ := os.ReadFile("./config.json")
	json.Unmarshal(file, &cnfg)
	//--------------------------------------------------------------------
	//Modbus TCP server
	fmt.Println("opening ModbusTCP server on localhost:502")
	serv := mbserver.NewServer()
	err := serv.ListenTCP(":502")
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
	// conn, err := net.Dial("tcp", ":8007")

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
	/*
		// for serial port
		//рабочий вариант, проверено по месту через сом порт
		mode := &serial.Mode{
			BaudRate: 4800,
			Parity:   serial.NoParity,
			DataBits: 8,
			StopBits: serial.OneStopBit,
		}
		port, err := serial.Open("COM10", mode)
		if err != nil {
			log.Fatal(err)
		}
		defer port.Close()
	*/
	//--------------------------------------------------------------------
	go iaswd(conn, &mutex)
	go FLAGs(conn, serv, &mutex)
	go SPs(conn, serv, &mutex)
	go APs(conn, serv, &mutex)
	go DACAs(conn, serv, &mutex)
	go WriteAll(conn, serv, &mutex)
	//--------------------------------------------------------------------
	for {
		time.Sleep(1 * time.Second)
	}
	//--------------------------------------------------------------------
}
