package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-dht"
)

func main() {
	// read DHT11 sensor data from pin 4, retrying 10 times in case of failure.
	// enable "boost GPIO performance" parameter, if your device is old as Raspberry BI 1
	// (this require root privileges)
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT22, 4, false, 10)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
