package main

import (
	"encoding/json"
	"flag"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const realtimeStatsQuery = "{\"emeter\":{\"get_realtime\":{}}}"

type EMeterReading struct {
	Name   string
	EMeter struct {
		Realtime struct {
			VoltageMv int `json:"voltage_mv"`
			CurrentMa int `json:"current_ma"`
			PowerMw   int `json:"power_mw"`
			TotalWh   int `json:"total_wh"`
			ErrCode   int `json:"err_code"`
		} `json:"get_realtime"`
	} `json:"emeter"`
	Time time.Time
}

func main() {
	targets := flag.String("target-urls", os.Getenv("TARGETS"), "comma delimited named target urls with name, e.g \"Socket1:192.168.0.100:9999\", (env: TARGETS)")
	influxUrl := flag.String("influx-url", os.Getenv("INFLUX_URL"), "InfluxDB url (env: INFLUX_URL)")
	bucketName := flag.String("bucket-name", os.Getenv("BUCKET_NAME"), "InfluxDB bucket name (env: BUCKET_NAME)")
	orgName := flag.String("org-name", os.Getenv("ORG_NAME"), "InfluxDB org name (env: ORG_NAME)")
	token := flag.String("token", os.Getenv("TOKEN"), "InfluxDB bucket access token (env: TOKEN)")
	flag.Parse()

	log.Println("Starting for targets: ", *targets)

	outChan := make(chan EMeterReading)

	tt := strings.Split(*targets, ",")
	for _, t := range tt {
		go getReading(t, outChan)
	}

	client := influxdb2.NewClient(*influxUrl, *token)
	influxWriter := client.WriteAPI(*orgName, *bucketName)
	defer client.Close()

	for r := range outChan {
		tags := map[string]string{
			"name": r.Name,
		}

		fields := map[string]interface{}{
			"voltage_mv": r.EMeter.Realtime.VoltageMv,
			"current_ma": r.EMeter.Realtime.CurrentMa,
			"power_mw":   r.EMeter.Realtime.PowerMw,
			"total_wh":   r.EMeter.Realtime.TotalWh,
		}

		point := influxdb2.NewPoint("power", tags, fields, r.Time)
		influxWriter.WritePoint(point)
	}
}

func getReading(target string, outChan chan EMeterReading) {
	index := strings.Index(target, ":")
	addr := target[index+1:]
	name := target[:index]
	requestBody := encode(realtimeStatsQuery)
	respBytes := make([]byte, 2048)
	for {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err != nil {
			log.Println("Failed to connect:", err)
			log.Println("Trying reset the connection...")
			time.Sleep(2 * time.Second)
		} else {
			for {
				_, err := conn.Write(requestBody)
				if err != nil {
					log.Println("Error writing to server:", err)
					log.Println("Trying reset the connection...")
					break
				}
				nBytes, err := conn.Read(respBytes)
				if err != nil {
					log.Println("Error reading response:", err)
					log.Println("Trying reset the connection...")
					break
				}

				response := decode(respBytes[4:nBytes])
				var reading EMeterReading

				err = json.Unmarshal(response, &reading)
				if err != nil {
					log.Printf("error unmarshalling response: %v, data: '%v'", err, string(decode(respBytes)))
					break
				}

				reading.Time = time.Now()
				reading.Name = name
				outChan <- reading

				time.Sleep(1 * time.Second)
			}
			time.Sleep(10 * time.Second)
		}
	}
}

func encode(s string) []byte {
	key := 171
	result := []byte{'\x00', '\x00', '\x00', '\x1e'}

	for _, v := range s {
		a := key ^ int(v)
		key = a
		result = append(result, byte(a))
	}

	return result[:len(s)+4]
}

func decode(s []byte) []byte {
	key := 171
	result := make([]byte, 0)

	for _, v := range s {
		a := key ^ int(v)
		key = int(v)
		result = append(result, byte(a))
	}
	return result
}
