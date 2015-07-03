## DHTxx temperature and humidity sensors

DHT11 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT11 (1).pdf)), DHT22 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT22.pdf)) sensors are quite popular among Arduino, Raspberry PI and their counterparts developers (here you will find comparision [DHT11 vs DHT22](https://raw.github.com/d2r2/go-dht/master/docs/dht.pdf)):
![dht11 and dht22](https://raw.github.com/d2r2/go-dht/master/docs/dht11_dht22.jpg)

They are cheap and affordable. So, here is a code written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts (has tested on Raspberry PI/Banana PI), which give you at the output temperature and humidity values (making all necessary signal handling behind the scenes).


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
The only thing which distinguish one from another - "retry count" parameter as additinal argument in ```ReadAndRetryDHTxx(...)```.
So, it's highly recomended to utilize ```ReadAndRetryDHTxx(...)``` with "retry count" not less than 7, since sensor asynchronouse protocol is not very stable causing errors time to time. Each additinal retry attempt spends 1.5-2 seconds (because according to specification before repeated attempt you should wait 1-2 seconds).

This functionality works not only with Raspberry PI, but with counterparts as well (tested with Raspberry PI and Banana PI). It will works with any Raspberry PI clone, which support Kernel SPI bus, but you should in advance make SPI bus device available in /dev/ list.

## License

Go-dht is licensed inder MIT License.

