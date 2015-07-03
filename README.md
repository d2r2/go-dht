## DHTxx temperature and humidity sensors

DHT11, DHT22 sensors are quite popular among Adruiono, Raspbery PI and their counterparts developers.
They are cheap and affordable. So here is a code which give you at the output temprature and humidity (make all necessary signal processing inside).

## Golang usage

```go
func main() {
	// read DHT11 sensor data from pin 4, retrying 10 times in case of failure
	err, temperature, humidity := dht.ReadAndRetryDHTxx(dht.DHT11, 4, 10)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%%\n", temperature, humidity)
}
```

## Getting help

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-dht/dht)

## Installation

```bash
$ go get github.com/d2r2/go-dht/dht
```

## Quick tutorial

There are two functions you could use: ```ReadDHTxx(...)``` and ```ReadAndRetryDHTxx(...)```.
They both do exactly same things - activate sensor and read and decode temperature and humidity.
Retry parameter - the only thing which distinguish one from another.
Nonetheless it's highly recomended to utilize ```ReadAndRetryDHTxx(...)```, since sensor asynchronouse protocol is not very stable causing errors time to time.

## License

Go-dht is licensed inder MIT License.

