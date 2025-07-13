package main

import (
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"net"
	"os"
	"sync"
	"time"

	"github.com/tbrandon/mbserver"
	"gopkg.in/yaml.v3"
)

// XOR всех байтов строки
func crc8(s string) byte {
	var ch byte
	for i := 0; i < len(s); i++ {
		ch ^= s[i]
	}
	return ch
}

// формируем строку в формате NMEA 0183
func nmea0183(data string) []byte {
	checksum := crc8(data)
	s := fmt.Sprintf("$%s*%02X\r\n", data, checksum)
	return []byte(s)
}

// собираем float из двух регистров
func Float32frombytes(a, b uint16) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint16(buf[0:2], a)
	binary.BigEndian.PutUint16(buf[2:4], b)
	bits := binary.BigEndian.Uint32(buf)
	float := math.Float32frombits(bits)
	return fmt.Sprintf("%.2f", float)
}

// ватчдог
func iaswd(q net.Conn, connMutex *sync.Mutex, cfg *Config) {
	ticker := time.NewTicker(time.Duration(cfg.WatchdogPeriod) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		msg := nmea0183("IAS WD")

		connMutex.Lock()
		_, err := q.Write(msg)
		connMutex.Unlock()

		if err != nil {
			log.Println("IAS watchdog write failed:", err)
			os.Exit(1) // Или передай в канал, если хочешь корректный shutdown
		}
	}
}

// загрузка конфига из yaml
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func loadTagConfig(path string) (*TagConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg TagConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// извлекаем данные из регистров в зависимости от типа данных
func extractValue(data []uint16, addr uint16, typ string) interface{} {
	switch typ {
	case "uint16":
		return data[addr]
	case "int16":
		return int16(data[addr])
	case "float32":
		bits := uint32(data[addr])<<16 | uint32(data[addr+1])
		return math.Float32frombits(bits)
	default:
		return nil
	}
}

// пишем все данные из карты в nmea периодично
func WriteAll(conn net.Conn, serv *mbserver.Server, regMutex *sync.Mutex, connMutex *sync.Mutex, tagCfg *TagConfig, cfg *Config) {
	ticker := time.NewTicker(time.Duration(cfg.WriteAllDelay) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		log.Println("Write all map: start")
		for _, tag := range tagCfg.Tags {
			current := make([]uint16, cfg.MapSize)
			// защищаем чтение карты мюьтексом
			regMutex.Lock()
			copy(current, serv.HoldingRegisters[:cfg.MapSize])
			regMutex.Unlock()

			value := extractValue(current, tag.Register, tag.Type)
			msg := fmt.Sprintf("ias,%s,%v", tag.Name, value)
			nmea := nmea0183(msg)
			// защищаем запись в терминальный сервер
			connMutex.Lock()
			_, err := conn.Write(nmea)
			connMutex.Unlock()

			if err != nil {
				log.Printf("Write error: %v", err)
				return
			}
		}
		log.Println("Write all map: finished")
	}
}

// смотрим изменения и кидаем в nmea
func MonitorTags(conn net.Conn, serv *mbserver.Server, regMutex *sync.Mutex, connMutex *sync.Mutex, tagCfg *TagConfig, cfg *Config) {
	// создаем массив для старых значений и тикер
	old := make([]uint16, cfg.MapSize)
	ticker := time.NewTicker(time.Duration(cfg.PoolingDelay) * time.Second)
	defer ticker.Stop()
	// итерируемся по тикеру
	for range ticker.C {
		current := make([]uint16, cfg.MapSize)
		// защищаем чтение карты мюьтексом
		regMutex.Lock()
		copy(current, serv.HoldingRegisters[:cfg.MapSize])
		regMutex.Unlock()

		for _, tag := range tagCfg.Tags {
			length := tag.Length
			if length == 0 {
				length = 1
			}
			// чекаем измененные значения
			changed := false
			for i := uint16(0); i < length; i++ {
				if current[tag.Register+i] != old[tag.Register+i] {
					changed = true
					break
				}
			}
			// если данные изменились
			if changed {
				value := extractValue(current, tag.Register, tag.Type)
				msg := fmt.Sprintf("ias,%s,%v", tag.Name, value)
				nmea := nmea0183(msg)
				// защищаем запись в терминальный сервер
				connMutex.Lock()
				_, err := conn.Write(nmea)
				connMutex.Unlock()

				if err != nil {
					log.Printf("Write error: %v", err)
					return
				}

				log.Printf("Sent NMEA: %s", msg)

				// обновляем old
				for i := uint16(0); i < length; i++ {
					old[tag.Register+i] = current[tag.Register+i]
				}

				time.Sleep(5 * time.Millisecond)
			}
		}
	}
}
