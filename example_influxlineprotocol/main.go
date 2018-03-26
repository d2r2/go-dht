package main

import (
	"flag"
	"fmt"
	"github.com/d2r2/go-dht"
	"math"
	"os"
	"time"
)

var (
	pin           int
	stype         string
	boostPerfFlag bool
)

func init() {
	flag.IntVar(&pin, "pin", 4, "pin")
	flag.StringVar(&stype, "sensor-type", "dht22", "sensor type (dht22, dht11)")
	flag.BoolVar(&boostPerfFlag, "boost", false, "boost performance")
}

func main() {
	flag.Parse()
	var sensorType dht.SensorType

	if stype == "dht22" || stype == "am2302" {
		sensorType = dht.DHT22
	} else if stype == "dht11" {
		sensorType = dht.DHT11
	}

	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(sensorType, pin, boostPerfFlag, 10)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)\n",
	//	sensorType, temperature, humidity, retried)
	// calculate vpd
	// J. Win. (https://physics.stackexchange.com/users/1680/j-win),
	// How can I calculate Vapor Pressure Deficit from Temperature and Relative Humidity?,
	// URL (version: 2011-02-03): https://physics.stackexchange.com/q/4553
	temperature64 := float64(temperature)
	humidity64 := float64(humidity)

	es := 0.6108 * math.Exp(17.27*temperature64/(temperature64+237.3))
	ea := humidity64 / 100 * es

	// this equation returns a negative value (in kPa), which while technically correct,
	// is invalid in this case because we are talking about a deficit.
	vpd := (ea - es) * -1

	str := fmt.Sprintf("%s temperature=%f,humidity=%f,vpd=%f,retried=%d %d",
		stype, temperature, humidity, vpd, retried, time.Now().UnixNano())
	fmt.Fprintln(os.Stdout, str)
}
