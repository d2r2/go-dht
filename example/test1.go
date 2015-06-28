package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-dht/dht"
)

func main() {
	err, temp, hum := dht.ReadAndRetryDHTxx(dht.DHT11, 4, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Temperature = %v*C, Humidity = %v%%\n", temp, hum)
}
