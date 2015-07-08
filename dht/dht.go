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
	Value    byte
	Duration time.Duration
}

func dialDHTxxAndGetResponse(pin int, boostPerfFlag bool) ([]Pulse, error) {
	var arr *C.int32_t
	var arrLen C.int32_t
	var l []int32
	var boost C.int32_t = 0
	if boostPerfFlag {
		boost = 1
	}
	// return array: [pulse, duration, pulse, duration, ...]
	r := C.dial_DHTxx_and_read(4, boost, &arr, &arrLen)
	if r == -1 {
		return nil, fmt.Errorf("Error during call C.dial_DHTxx_and_read()")
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
		var value byte = 0
		if l[i*2] != 0 {
			value = 1
		}
		pulses[i] = Pulse{Value: value,
			Duration: time.Duration(l[i*2+1]) * time.Microsecond}
	}
	return pulses, nil
}

func decodeByte(pulses []Pulse, start int) (int, error) {
	if len(pulses)-start < 16 {
		return 0, fmt.Errorf("Can't decode byte, since range between "+
			"index and array length is less than 16: %d, %d", start, len(pulses))
	}
	var b int = 0
	for i := 0; i < 8; i++ {
		pulseL := pulses[start+i*2]
		pulseH := pulses[start+i*2+1]
		if pulseL.Value != 0 {
			return 0, fmt.Errorf("Low edge value expected at index %d", start+i*2)
		}
		if pulseH.Value == 0 {
			return 0, fmt.Errorf("High edge value expected at index %d", start+i*2+1)
		}
		const HIGH_DUR_MAX = (70 + (70 + 54)) / 2 * time.Microsecond
		// Calc average value between 24us (bit 0) and 70us (bit 1).
		// Everything that less than this param is bit 0, bigger - bit 1.
		const HIGH_DUR_AVG = (24 + (70-24)/2) * time.Microsecond
		if pulseH.Duration > HIGH_DUR_MAX {
			return 0, fmt.Errorf("High edge value duration exceed "+
				"expected maximum amount in %v: %v", HIGH_DUR_MAX, pulseH.Duration)
		}
		if pulseH.Duration > HIGH_DUR_AVG {
			//fmt.Printf("bit %d is high\n", 7-i)
			b = b | (1 << uint(7-i))
		}
	}
	return b, nil
}

// Decode bunch of pulse read from DHTxx sensors.
// Use pdf description from /docs to read 5 bytes and
// convert them to temperature and humidity.
func decodeDHT11Pulses(pulses []Pulse) (temperature float32,
	humidity float32, err error) {
	if len(pulses) == 85 {
		pulses = pulses[3:]
	} else if len(pulses) == 84 {
		pulses = pulses[2:]
	} else if len(pulses) == 83 {
		pulses = pulses[1:]
	} else if len(pulses) != 82 {
		printPulseArrayForDebug(pulses)
		return -1, -1, fmt.Errorf("Can't decode pulse array received from "+
			"DHTxx sensor, since incorrect length: %d", len(pulses))
	}
	pulses = pulses[:80]
	// Decode humidity (integer part)
	humInt, err := decodeByte(pulses, 0)
	if err != nil {
		return -1, -1, err
	}
	// Decode humidity (decimal part)
	humDec, err := decodeByte(pulses, 16)
	if err != nil {
		return -1, -1, err
	}
	// Decode temperature (integer part)
	tempInt, err := decodeByte(pulses, 32)
	if err != nil {
		return -1, -1, err
	}
	// Decode temperature (decimal part)
	tempDec, err := decodeByte(pulses, 48)
	if err != nil {
		return -1, -1, err
	}
	// Decode control sum to verify all data received from sensor
	sum, err := decodeByte(pulses, 64)
	if err != nil {
		return -1, -1, err
	}
	// Produce data verification
	if byte(sum) != byte(humInt+humDec+tempInt+tempDec) {
		return -1, -1, fmt.Errorf("Control sum %d doesn't match %d (%d+%d+%d+%d)",
			sum, byte(humInt+humDec+tempInt+tempDec),
			humInt, humDec, tempInt, tempDec)
	}
	temperature = float32(tempInt)
	humidity = float32(humInt)
	if humidity > 100 {
		return -1, -1, fmt.Errorf("Humidity value exceed 100%: %v", humidity)
	}
	// Success
	return temperature, humidity, nil
}

func printPulseArrayForDebug(pulses []Pulse) {
	fmt.Printf("Pulse count %d:\n", len(pulses))
	for i, pulse := range pulses {
		fmt.Printf("\tpulse %3d: %v, %v\n", i, pulse.Value, pulse.Duration)
	}
}

// Send activation request to DHTxx sensor via 1-pin.
// Then decode pulses which was sent back with asynchronous
// protocol specific for DHTxx sensors.
func ReadDHTxx(sensorType SensorType, pin int,
	boostPerfFlag bool) (temperature float32, humidity float32, err error) {
	pulses, err := dialDHTxxAndGetResponse(pin, boostPerfFlag)
	if err != nil {
		return -1, -1, err
	}
	temp, hum, err := decodeDHT11Pulses(pulses)
	if err != nil {
		return -1, -1, err
	}
	return temp, hum, nil
}

// Read temperature (in celcius) and humidity (in percents)
// from DHTxx sensors. Retry n times in case of failure.
func ReadDHTxxWithRetry(sensorType SensorType, pin int, retry int,
	boostPerfFlag bool) (temperature float32, humidity float32, retried int, err error) {
	retried = 0
	for {
		temp, hum, err := ReadDHTxx(sensorType, pin, boostPerfFlag)
		if err != nil {
			log.Println(err)
			if retry > 0 {
				retry--
				retried++
				// Sleep before new attempt
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			return -1, -1, retried, err
		}
		if retried > 0 {
			log.Printf("Success! Retried %d times\n", retried)
		}
		return temp, hum, retried, nil
	}
}
