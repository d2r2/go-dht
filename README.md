## DHTxx temperature and humidity sensors

DHT11 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT11 (1).pdf)) and DHT22 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT22.pdf)) sensors are quite popular among Arduino, Raspberry PI and their counterparts developers (here you will find comparision [DHT11 vs DHT22](https://raw.github.com/d2r2/go-dht/master/docs/dht.pdf)):
![dht11 and dht22](https://raw.github.com/d2r2/go-dht/master/docs/dht11_dht22.jpg)

They are cheap enough and affordable. So, here is a code written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts, which gives you at the output temperature and humidity values (making all necessary signal processing via their own 1-wire bus protocol behind the scene).

## Compatibility

Tested on Raspberry PI 1 (model B) and Banana PI (model M1).

## Golang usage

```go
func main() {
	// Read DHT11 sensor data from pin 4, retrying 10 times in case of failure.
	// You may enable "boost GPIO performance" parameter, if your device is old
	// as Raspberry PI 1 (this will require root privileges). You can switch off
	// "boost GPIO performance" parameter for old devices, but it may increase
	// retry attempts. Play with this parameter.
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(dht.DHT11, 4, true, 10)
	if err != nil {
		log.Fatal(err)
	}
	// Print temperature and humidity
	fmt.Printf("Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
		temperature, humidity, retried)
}
```

## Getting help

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-dht).

For detailed explanation read great article "[Golang with Raspberry Pi : Read RH and Temperature from DHT22 or AM2302](https://skylabin.wordpress.com/2015/09/18/golang-with-raspberry-pi-read-rh-and-temperature-from-dht22-or-am2302)" written by Joseph Mathew. Thanks Joseph!

## Installation

```bash
$ go get -u github.com/d2r2/go-dht
```

## Quick tutorial

There are two functions you could use: ```ReadDHTxx(...)``` and ```ReadDHTxxWithRetry(...)```.
They both do exactly same thing - activate sensor then read and decode temperature and humidity values.
The only thing which distinguish one from another - "retry count" parameter as additinal argument in ```ReadDHTxxWithRetry(...)```.
So, it's highly recomended to utilize ```ReadDHTxxWithRetry(...)``` with "retry count" not less than 7, since sensor asynchronouse protocol is not very stable causing errors time to time. Each additinal retry attempt takes 1.5-2 seconds (according to specification before repeated attempt you should wait 1-2 seconds).

This functionality works not only with Raspberry PI, but with counterparts as well (tested with Raspberry PI and Banana PI).

> Note: If you enable "boost GPIO performance" parameter, application should run with root privileges, since C code inside requires this. In most cases it is sufficient to add "sudo -E" before "go run ...".

> Note: This package does not require any external C code or library.

## License

Go-dht is licensed under MIT License.

