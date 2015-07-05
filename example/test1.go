package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-dht/dht"
)

func main() {
	// read DHT11 sensor data from pin 4, retrying 10 times in case of failure
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT11, 4, 10)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
