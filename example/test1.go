package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-dht"
)

func main() {
	// read DHTxx sensor data from pin 4, retrying 10 times in case of failure
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT22, 4, 10, false)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
