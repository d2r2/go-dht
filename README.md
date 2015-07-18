## DHTxx temperature and humidity sensors

DHT11 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT11 (1).pdf)) and DHT22 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT22.pdf)) sensors are quite popular among Arduino, Raspberry PI and their counterparts developers (here you will find comparision [DHT11 vs DHT22](https://raw.github.com/d2r2/go-dht/master/docs/dht.pdf)):
![dht11 and dht22](https://raw.github.com/d2r2/go-dht/master/docs/dht11_dht22.jpg)

They are cheap and affordable. So, here is a code written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts (tested on Raspberry PI/Banana PI), which gives you at the output temperature and humidity values (making all necessary signal processing behind the scenes).


## Golang usage

```go
func main() {
	// read DHT11 sensor data from pin 4, retrying 10 times in case of failure.
	// enable "boost GPIO performance" parameter, if your device is old as Raspberry BI 1 (this
	// require root privileges)
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT11, 4, true, 10)
	if err != nil {
		log.Fatal(err)
	}
	// print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
```

## Getting help

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-dht)

## Installation

```bash
$ go get -u github.com/d2r2/go-dht
```

## Quick tutorial

There are two functions you could use: ```ReadDHTxx(...)``` and ```ReadDHTxxWithRetry(...)```.
They both do exactly same thing - activate sensor and read and decode temperature and humidity values.
The only thing which distinguish one from another - "retry count" parameter as additinal argument in ```ReadDHTxxWithRetry(...)```.
So, it's highly recomended to utilize ```ReadDHTxxWithRetry(...)``` with "retry count" not less than 7, since sensor asynchronouse protocol is not very stable causing errors time to time. Each additinal retry attempt takes 1.5-2 seconds (according to specification before repeated attempt you should wait 1-2 seconds).

This functionality works not only with Raspberry PI, but with counterparts as well (tested with Raspberry PI and Banana PI). It will works with any Raspberry PI clone, which support Kernel SPI bus, but you should in advance make SPI bus device available in /dev/ list.

NOTE: If you enable "boost GPIO performance" parameter, application should run with root privileges, since C code inside requires it. In most cases it is sufficient to add "sudo -E" before "go run ...".

## License

Go-dht is licensed under MIT License.

