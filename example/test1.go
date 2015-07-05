package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-dht/dht"
)

func main() {
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT11, 4, 10)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
