package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/andschneider/goqtt"
	"github.com/andschneider/goqtt/packets"
	influxdb2 "github.com/influxdata/influxdb-client-go"
)

// config contains the necessary information to create clients for both
// Influx and goqtt. It also holds an InfluxDB client, which after a
// successful connection can be used to write data.
type config struct {
	// address of influxdb
	influxDB string
	// influxdb bucket name
	influxBucket string
	// influxdb client
	influxClient influxdb2.Client

	// address of MQTT broker
	mqttServer string
	// MQTT topic to subscribe to
	mqttTopic string
}

// loadConfig loads in the required configuration from environment variables.
// If a variable isn't set the program will exit with an exit code of 1.
func loadConfig() *config {
	defaultEnv := func(n string) string {
		e := os.Getenv(n)
		if e == "" {
			log.Printf("must set %s env variable\n", n)
			os.Exit(1)
		}
		return e
	}
	cfg := &config{
		influxDB:     defaultEnv("INFLUX_HOST"),
		influxBucket: defaultEnv("INFLUX_BUCKET"),
		mqttServer:   defaultEnv("MQTT_HOST"),
		mqttTopic:    defaultEnv("MQTT_TOPIC"),
	}
	return cfg
}

// reading is a struct representing the expected sensor reading data in
// the MQTT message. It is expected to a be a JSON.
type reading struct {
	Moisture    int     `json:"moisture"`
	Temperature float32 `json:"temperature"`
	Sid         string  `json:"sid"`
}

// writeData writes the data to influx using the blocking API.
// Might switch to non-blocking later.
func (c *config) writeData(line string) {
	log.Printf("writing line: %s", line)
	writeApi := c.influxClient.WriteAPIBlocking("", c.influxBucket)
	err := writeApi.WriteRecord(context.Background(), line)
	if err != nil {
		log.Printf("write error: %s\n", err.Error())
	}
}

// handleMessage unmarshalls the MQTT message and saves it to influx.
func (c *config) handleMessage(m *packets.PublishPacket) {
	log.Printf("received message: '%s' from topic: '%s'\n", string(m.Message), m.Topic)
	// mes := []byte(`{"moisture": 588, "temperature": 26.39, "sid": "sensor1"}`)
	r := reading{}
	if err := json.Unmarshal(m.Message, &r); err != nil {
		log.Printf("could not unmarshal json data: %v", err)
	}

	// save data to influx
	moist := fmt.Sprintf("moisture,unit=capacitance,sensor=%s avg=%d", r.Sid, r.Moisture)
	//log.Printf("moisture: %s\n", moist)
	c.writeData(moist)

	temp := fmt.Sprintf("temperature,unit=celsius,sensor=%s avg=%f", r.Sid, r.Temperature)
	//log.Printf("temperature: %s\n", temp)
	c.writeData(temp)
}

func main() {
	cfg := loadConfig()

	// connect to MQTT broker
	log.Println("connecting to MQTT")
	mqttClient := goqtt.NewClient(cfg.mqttServer, goqtt.Topic(cfg.mqttTopic))

	err := mqttClient.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer mqttClient.Disconnect()

	// setup influx connection
	log.Println("connecting to Influx")
	cfg.influxClient = influxdb2.NewClient(cfg.influxDB, "")
	defer cfg.influxClient.Close()

	// Subscribe to MQTT topic
	err = mqttClient.Subscribe()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("waiting for messages...")
	for {
		m, err := mqttClient.ReadLoop()
		if err != nil {
			log.Printf("error: read loop: %v\n", err)
		}
		if m != nil {
			cfg.handleMessage(m)
		}
	}
}
