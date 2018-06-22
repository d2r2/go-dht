DHTxx temperature and humidity sensors
======================================

[![Build Status](https://travis-ci.org/d2r2/go-dht.svg?branch=master)](https://travis-ci.org/d2r2/go-dht)
[![Go Report Card](https://goreportcard.com/badge/github.com/d2r2/go-dht)](https://goreportcard.com/report/github.com/d2r2/go-dht)
[![GoDoc](https://godoc.org/github.com/d2r2/go-dht?status.svg)](https://godoc.org/github.com/d2r2/go-dht)
[![MIT License](http://img.shields.io/badge/License-MIT-yellow.svg)](./LICENSE)
<!--
[![Coverage Status](https://coveralls.io/repos/d2r2/go-dht/badge.svg?branch=master)](https://coveralls.io/r/d2r2/go-dht?branch=master)
-->


DHT11 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT11-2.pdf)) and DHT22 ([pdf reference](https://raw.github.com/d2r2/go-dht/master/docs/DHT22.pdf)) sensors are quite popular among Arduino, Raspberry PI and their counterparts developers (here you will find comparision [DHT11 vs DHT22](https://raw.github.com/d2r2/go-dht/master/docs/dht.pdf)):
![dht11 and dht22](https://raw.github.com/d2r2/go-dht/master/docs/dht11_dht22.jpg)

They are cheap enough and affordable. So, here is a code written in [Go programming language](https://golang.org/) for Raspberry PI and counterparts, which gives you at the output temperature and humidity values (making all necessary signal processing via their own 1-wire bus protocol behind the scene).


Technology overview
-------------------

There are 2 methods how we can drive such devices which requre special pins switch from low to high level and back (employing specific "1-wire protocol" described in pdf documentation):
1) First approach implies to work on the most lower layer to handle pins via GPIO chip registers using linux "memory mapped" device (/dev/mem). This approach is most reliable (until you move to other RPI clone) and fastest with regard to the transmission speed. Disadvantage of this method is explained by the fact that each RPI-clone have their own GPIO registers set to drive device GPIO pins.
2) Second option implies to access GPIO pins via special layer based on linux "device tree" approach (/sys/class/gpio/... virtual file system), which translate such operations to direct register writes and reads described in 1st approach. In some sence it is more compatible when you move from original Raspberry PI to RPI-clones, but may have some issues in stability of specific implementations. As it found some clones don't implement this layer at all from the box (Beaglebone for instance). 

So, here I'm using second approach.

Compatibility
-------------

Tested on Raspberry PI 1 (model B), Banana PI (model M1), Orange PI One.

Golang usage
------------

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

Installation
------------

```bash
$ go get -u github.com/d2r2/go-dht
```

Quick start
-----------

There are two functions you could use: ```ReadDHTxx(...)``` and ```ReadDHTxxWithRetry(...)```.
They both do exactly same thing - activate sensor then read and decode temperature and humidity values.
The only thing which distinguish one from another - "retry count" parameter as additinal argument in ```ReadDHTxxWithRetry(...)```.
So, it's highly recomended to utilize ```ReadDHTxxWithRetry(...)``` with "retry count" not less than 7, since sensor asynchronouse protocol is not very stable causing errors time to time. Each additinal retry attempt takes 1.5-2 seconds (according to specification before repeated attempt you should wait 1-2 seconds).

This functionality works not only with Raspberry PI, but with counterparts as well (tested with Raspberry PI and Banana PI).

> Note: If you enable "boost GPIO performance" parameter, application should run with root privileges, since C code inside requires this. In most cases it is sufficient to add "sudo -E" before "go run ...".

> Note: This package does not have dependency on any sensor-specific 3-rd party C-code or library.

Tutorial
--------

Library comprised of 2 parts: low level to send queries and read raw data from sensor written in C-code and front end functions with decoding raw data in Golang.

Originally attempt was made to write whole library in Golang, but during debugging it was found that Garbage Collector (GC) "stop the world" issue in early version of Golang sometimes freeze library in the middle of sensor reading process, which lead to unpredictable mistakes when some signals from sensor are missing.  Starting from Go 1.5 version GC behaviour was improved significantly, but original design left as is since it has been tested and works reliably in most cases.

To install library on your Raspberry PI device you should execute console command `go get -u github.com/d2r2/go-dht` to download and install/update package to you device `$GOPATH/src` path.

You may start from simple test with DHTxx sensor using `./example/test1.go` application which will interact with the sensor connected to physical pin 7 (which correspond to GPIO4 pin-out).

Also you can use cross compile technique, to build ARM application from x86/64bit system. For this your should install GCC tool-chain for ARM target platform. So, your x86/64bit linux system should have specific gcc compiler installed: in case of Debian or Ubuntu `arm-linux-gnueabi-gcc` (in case of Arch linux `arm-linux-gnueabihf-gcc`).
After all, for instance, for cross compiling test application "./examples/example1/example1.go" to ARM target platform in Ubuntu/Debian you should run `CC=arm-linux-gnueabi-gcc CGO_ENABLED=1 GOOS=linux GOARCH=arm GOARM=6 go build ./examples/example1/example1.go`.

GoDoc [documentation](http://godoc.org/github.com/d2r2/go-dht).

For detailed explanation read great article "[Golang with Raspberry Pi : Read RH and Temperature from DHT22 or AM2302](https://skylabin.wordpress.com/2015/09/18/golang-with-raspberry-pi-read-rh-and-temperature-from-dht22-or-am2302)" written by Joseph Mathew. Thanks Joseph!

Contribute authors
------------------

* Joseph Mathew (https://skylabin.wordpress.com/)
* Alex Zhang ([ztc1997](https://github.com/ztc1997))
* Andy Brown ([andybrown668](https://github.com/andybrown668))
* Gareth Dunstone ([gdunstone](https://github.com/gdunstone))

Contact
-------

Please use [Github issue tracker](https://github.com/d2r2/go-dht/issues) for filing bugs or feature requests.

License
-------

Go-dht is licensed under MIT License.
