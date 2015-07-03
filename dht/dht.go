package dht

// #include "dht.go.h"
// #cgo LDFLAGS: -lrt
import "C"

import (
	"fmt"
	"log"
	"reflect"
	"time"
	"unsafe"
)

type SensorType int

const (
	DHT11 SensorType = iota + 1
	DHT22
)

type Pulse struct {
	Value    bool
	Duration time.Duration
}

func dialDHTxxAndGetResponse(pin int) (error, []Pulse) {
	var arr *C.int32_t
	var arrLen C.int32_t
	var l []int32
	// return array: [pulse, duration, pulse, duration, ...]
	r := C.dial_DHTxx_and_read(4, &arr, &arrLen)
	if r == -1 {
		return fmt.Errorf("Error found in call to C.dial_DHTxx_and_read(...)"), nil
	}
	defer C.free(unsafe.Pointer(arr))
	h := (*reflect.SliceHeader)(unsafe.Pointer(&l))
	h.Data = uintptr(unsafe.Pointer(arr))
	h.Len = int(arrLen)
	h.Cap = int(arrLen)
	pulses := make([]Pulse, len(l)/2)
	// convert original int array ([pulse, duration, pulse, duration, ...])
	// to Pulse struct array
	for i := 0; i < len(l)/2; i++ {
		pulses[i] = Pulse{Value: l[i*2] != 0,
			Duration: time.Duration(l[i*2+1]) * time.Microsecond}
	}
	time.
	return nil, pulses
}

func decodeByte(pulses []Pulse, start int) (error, int) {
	if len(pulses)-start < 16 {
		return fmt.Errorf("Can't decode byte, since range between "+
			"index and array length is less than 16: %d, %d", start, len(pulses)), 0
	}
	var b int = 0
	for i := 0; i < 8; i++ {
		pulseL := pulses[start+i*2]
		pulseH := pulses[start+i*2+1]
		if pulseL.Value != false {
			return fmt.Errorf("Low edge value expected at index %d", start+i*2), 0
		}
		if pulseH.Value != true {
			return fmt.Errorf("High edge value expected at index %d", start+i*2+1), 0
		}
		const HIGH_DUR_MAX = (70 + (70 + 54)) / 2 * time.Microsecond
		// Calc average value between 24us (bit 0) and 70us (bit 1).
		// Everything that less than this param is bit 0, bigger - bit 1.
		const HIGH_DUR_AVG = (24 + (70-24)/2) * time.Microsecond
		if pulseH.Duration > HIGH_DUR_MAX {
			return fmt.Errorf("High edge value duration exceed "+
				"expected maximum amount in %v: %v", HIGH_DUR_MAX, pulseH.Duration), 0
		}
		if pulseH.Duration > HIGH_DUR_AVG {
			//fmt.Printf("bit %d is high\n", 7-i)
			b = b | (1 << uint(7-i))
		}
	}
	return nil, b
}

// Decode bunch of pulse read from DHTxx sensors.
// Use pdf description from /docs to read 5 bytes and
// convert them to temperature and humidity.
func decodeDHT11Pulses(pulses []Pulse) (err error, temp float32, hum float32) {
	if len(pulses) == 84 {
		pulses = pulses[2:82]
	} else if len(pulses) == 83 {
		pulses = pulses[1:81]
	} else if len(pulses) != 82 {
		return fmt.Errorf("Can't decode pulse array, since incorrect length: %d",
			len(pulses)), -1, -1
	}
	pulses = pulses[:80]
	// Decode humidity (integer part)
	err, humInt := decodeByte(pulses, 0)
	if err != nil {
		return err, -1, -1
	}
	// Decode humidity (decimal part)
	err, humDec := decodeByte(pulses, 16)
	if err != nil {
		return err, -1, -1
	}
	// Decode temperature (integer part)
	err, tempInt := decodeByte(pulses, 32)
	if err != nil {
		return err, -1, -1
	}
	// Decode temperature (decimal part)
	err, tempDec := decodeByte(pulses, 48)
	if err != nil {
		return err, -1, -1
	}
	// Decode control sum to verify all data received from sensor
	err, sum := decodeByte(pulses, 64)
	if err != nil {
		return err, -1, -1
	}
	// Produce data verification
	if byte(sum) != byte(humInt+humDec+tempInt+tempDec) {
		return fmt.Errorf("Control sum %d doesn't match %d+%d+%d+%d=%d\n",
			sum, humInt, humDec, tempInt, tempDec,
			byte(humInt+humDec+tempInt+tempDec)), -1, -1
	}
	temp = float32(tempInt)
	hum = float32(humInt)
	if hum > 100 {
		return fmt.Errorf("Humidity value exceed 100%: %v", hum), -1, -1
	}
	// Success
	return nil, temp, hum
}

// Send activation request to DHTxx sensor via 1-pin.
// Then decode pulses which was sent back with asynchronous
// protocol specific for DHTxx sensors.
func ReadDHTxx(sensorType SensorType, pin int) (err error,
	temperature float32, humidity float32) {
	err, pulses := dialDHTxxAndGetResponse(pin)
	if err != nil {
		return err, -1, -1
	}
	// Uncomment next block for debug purpose.
	/*
		for i, pulse := range pulses {
			fmt.Printf("pulse %d: %v, %v\n", i, pulse.Duration, pulse.Value)
		}
	*/
	err, temp, hum := decodeDHT11Pulses(pulses)
	if err != nil {
		return err, -1, -1
	}
	return nil, temp, hum
}

// Read temperature (in celcius) and humidity (in percents)
// from DHTxx sensors. Retry n times in case of failure.
func ReadAndRetryDHTxx(sensorType SensorType, pin int, retry int) (err error,
	temperature float32, humidity float32) {
	var retryUsed int = 0
	for {
		err, temp, hum := ReadDHTxx(sensorType, pin)
		if err != nil {
			log.Println(err)
			if retry > 0 {
				retry--
				retryUsed++
				// Sleep before new attempt
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			return err, -1, -1
		}
		if retryUsed > 0 {
			fmt.Printf("Has retried %d times\n", retryUsed)
		}
		return nil, temp, hum
	}
}
