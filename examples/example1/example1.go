package main

import (
	"github.com/d2r2/go-dht"

	logger "github.com/d2r2/go-logger"
)

var lg = logger.NewPackageLogger("main",
	logger.DebugLevel,
	// logger.InfoLevel,
)

func main() {
	defer logger.FinalizeLogger()
	// Uncomment next line to suppress verbose output
	logger.ChangePackageLogLevel("dht", logger.InfoLevel)
	sensorType := dht.DHT11
	// Read DHT11 sensor data from pin 4, retrying 10 times in case of failure.
	// You may enable "boost GPIO performance" parameter, if your device is old
	// as Raspberry PI 1 (this will require root privileges). You can switch off
	// "boost GPIO performance" parameter for old devices, but it may increase
	// retry attempts. Play with this parameter.
	temperature, humidity, retried, err :=
		dht.ReadDHTxxWithRetry(sensorType, 4, false, 10)
	if err != nil {
		lg.Fatal(err)
	}
	// print temperature and humidity
	lg.Infof("Sensor = %v: Temperature = %v*C, Humidity = %v%% (retried %d times)",
		sensorType, temperature, humidity, retried)
}
