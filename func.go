package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/tbrandon/mbserver"
)

func crc8(s string) byte {
	var ch byte
	if len(s) == 0 {
		return 0
	}

	for i := 1; i < len(s); i++ {
		ch = ch ^ s[i]
	}
	return ch
}

func nmea0183(data string) []byte {
	ss := make([]byte, 0, 100)
	ss = append(ss, []byte("$")...)
	ss = append(ss, []byte(data)...)
	ss = append(ss, []byte("*")...)
	ss = append(ss, []byte(strconv.FormatInt(int64(crc8(data)), 16))...)
	ss = append(ss, []byte("\r\n")...)
	return ss
}

func Float32frombytes(a, b uint16) string {
	var sl []byte
	sl1 := make([]byte, 2)
	sl2 := make([]byte, 2)
	binary.BigEndian.PutUint16(sl1, a)
	binary.BigEndian.PutUint16(sl2, b)
	sl = append(sl, sl1...)
	sl = append(sl, sl2...)
	bits := binary.BigEndian.Uint32(sl)
	float := math.Float32frombits(bits)
	return fmt.Sprintf("%8.2f", float)
}

func iaswd(q net.Conn, mutex *sync.Mutex) {
	for {

		mutex.Lock()
		wdStr := "IAS WD"
		_, err := q.Write(nmea0183(wdStr))
		mutex.Unlock()

		if err != nil {
			log.Println("IAS watchdog ", err.Error())
			os.Exit(1)
		}
		time.Sleep(8 * time.Second)
	}
}

func FLAGs(conn net.Conn, serv *mbserver.Server, mutex *sync.Mutex) {
	oldTable := make([]uint16, 200)
	for {
		mutex.Lock()
		for i := 0; i < 195; i++ {
			if serv.HoldingRegisters[i+299] != oldTable[i] {
				w := FALAGA[i] + strconv.Itoa(int(serv.HoldingRegisters[i+299]))
				_, err := conn.Write(nmea0183(w))
				if err != nil {
					log.Println("Write to server failed:", err.Error())
					os.Exit(1)
				}
				log.Println("transmitted : " + w)
				oldTable[i] = serv.HoldingRegisters[i+299]
				time.Sleep(10 * time.Millisecond)
			}
		}
		mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}
func SPs(conn net.Conn, serv *mbserver.Server, mutex *sync.Mutex) {

	oldTable := make([]uint16, 250)
	for {
		mutex.Lock()
		for i := 0; i < 249; i++ {
			if serv.HoldingRegisters[i] != oldTable[i] {
				w := StatusPoint[i] + strconv.Itoa(int(serv.HoldingRegisters[i]))
				_, err := conn.Write(nmea0183(w))
				if err != nil {
					log.Println("Write to server failed:", err.Error())
					os.Exit(1)
				}
				log.Println("transmitted : " + w)
				oldTable[i] = serv.HoldingRegisters[i]
				time.Sleep(10 * time.Millisecond)
			}
		}
		mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}
func APs(conn net.Conn, serv *mbserver.Server, mutex *sync.Mutex) {
	oldTable := make([]uint16, 350)
	for {
		mutex.Lock()
		for i := 0; i < 350; i += 2 {
			if serv.HoldingRegisters[i+598] != oldTable[i] || serv.HoldingRegisters[i+599] != oldTable[i+1] {
				w := AnalogPoint[i/2] + Float32frombytes(serv.HoldingRegisters[i+598], serv.HoldingRegisters[i+599])
				_, err := conn.Write(nmea0183(w))
				if err != nil {
					log.Println("Write to server failed:", err.Error())
					os.Exit(1)
				}
				log.Println("transmitted : " + w)
				oldTable[i] = serv.HoldingRegisters[i+598]
				oldTable[i+1] = serv.HoldingRegisters[i+599]
				time.Sleep(10 * time.Millisecond)
			}
		}
		mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}
func DACAs(conn net.Conn, serv *mbserver.Server, mutex *sync.Mutex) {
	oldTable := make([]uint16, 250)
	for {
		mutex.Lock()
		for i := 0; i < 249; i += 2 {
			if serv.HoldingRegisters[i+1198] != oldTable[i] || serv.HoldingRegisters[i+1199] != oldTable[i+1] {
				w := DACA[i/2] + Float32frombytes(serv.HoldingRegisters[i+1198], serv.HoldingRegisters[i+1199])
				_, err := conn.Write(nmea0183(w))
				if err != nil {
					log.Println("Write to server failed:", err.Error())
					os.Exit(1)
				}
				log.Println("transmitted : " + w)
				oldTable[i] = serv.HoldingRegisters[i+1198]
				oldTable[i+1] = serv.HoldingRegisters[i+1199]
				time.Sleep(10 * time.Millisecond)
			}
		}
		mutex.Unlock()
		time.Sleep(1 * time.Second)
	}
}

func WriteAll(conn net.Conn, serv *mbserver.Server, mutex *sync.Mutex) {
	for {
		ClearScreen()
		mutex.Lock()
		log.Println("start to transmit All map")
		for i := 0; i < 195; i++ {
			w := FALAGA[i] + strconv.Itoa(int(serv.HoldingRegisters[i+299]))
			_, err := conn.Write(nmea0183(w))
			// fmt.Println("flaga area: ", w)
			if err != nil {
				log.Println("Write to server failed:", err.Error())
				os.Exit(1)
			}
			time.Sleep(time.Duration(cnfg.PoolingDelay) * time.Millisecond)
		}
		// time.Sleep(1 * time.Second)

		for i := 0; i < 249; i++ {
			w := StatusPoint[i] + strconv.Itoa(int(serv.HoldingRegisters[i]))
			_, err := conn.Write(nmea0183(w))
			// fmt.Println("SP area: ", w)
			if err != nil {
				log.Println("Write to server failed:", err.Error())
				os.Exit(1)
			}
			time.Sleep(time.Duration(cnfg.PoolingDelay) * time.Millisecond)
		}
		// time.Sleep(1 * time.Second)

		for i := 0; i < 346; i += 2 {
			w := AnalogPoint[i/2] + Float32frombytes(serv.HoldingRegisters[i+598], serv.HoldingRegisters[i+599])
			_, err := conn.Write(nmea0183(w))
			// fmt.Println("AP area: ", w)
			if err != nil {
				log.Println("Write to server failed:", err.Error())
				os.Exit(1)
			}
			time.Sleep(time.Duration(cnfg.PoolingDelay) * time.Millisecond)
		}
		// time.Sleep(1 * time.Second)

		for i := 0; i < 248; i += 2 {
			w := DACA[i/2] + Float32frombytes(serv.HoldingRegisters[i+1198], serv.HoldingRegisters[i+1199])
			_, err := conn.Write(nmea0183(w))
			// fmt.Println("flaga DACA: ", w)
			if err != nil {
				log.Println("Write to server failed:", err.Error())
				os.Exit(1)
			}
			time.Sleep(time.Duration(cnfg.PoolingDelay) * time.Millisecond)
		}
		log.Println("transmitted All map")
		mutex.Unlock()
		time.Sleep(time.Duration(cnfg.WriteAllDelay) * time.Second)

	}

}

func ClearScreen() {

	if runtime.GOOS == "windows" {
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	} else {
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	fmt.Println("\033[32mVDR: converter ModbusTCP to NMEA0183\033[0m")       //зеленый
	fmt.Println("\033[32mModbusTCP server opened on localhost:502\033[0m\n") //зеленый

}
