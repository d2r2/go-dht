package dht

// #include "dht.go.h"
// #cgo LDFLAGS: -lrt
import "C"

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"
	"unsafe"

	"github.com/d2r2/go-shell"
	"github.com/davecgh/go-spew/spew"
)

type SensorType int

// Implement Stringer interface.
func (this SensorType) String() string {
	if this == DHT11 {
		return "DHT11"
	} else if this == DHT22 {
		return "DHT22"
	} else if this == AM2302 {
		return "AM2302"
	} else {
		return "!!! unknown !!!"
	}
}

const (
	// Most populare sensor
	DHT11 SensorType = iota + 1
	// More expensive and precise than DHT11
	DHT22
	// Aka DHT22
	AM2302 = DHT22
)

// Keep pulse state with how long it lasted.
type Pulse struct {
	Value    byte
	Duration time.Duration
}

// Activate sensor and get back bunch of pulses for further decoding.
// C function call wrapper.
func dialDHTxxAndGetResponse(pin int, boostPerfFlag bool) ([]Pulse, error) {
	var arr *C.int32_t
	var arrLen C.int32_t
	var list []int32
	var boost C.int32_t = 0
	var err2 *C.Error
	if boostPerfFlag {
		boost = 1
	}
	// Return array: [pulse, duration, pulse, duration, ...]
	r := C.dial_DHTxx_and_read(C.int32_t(pin), boost, &arr, &arrLen, &err2)
	if r == -1 {
		var err error
		if err2 != nil {
			msg := C.GoString(err2.message)
			err = errors.New(spew.Sprintf("Error during call C.dial_DHTxx_and_read(): %v", msg))
			C.free_error(err2)
		} else {
			err = errors.New(spew.Sprintf("Error during call C.dial_DHTxx_and_read()"))
		}
		return nil, err
	}
	defer C.free(unsafe.Pointer(arr))
	// Convert original C array arr to Go slice list
	h := (*reflect.SliceHeader)(unsafe.Pointer(&list))
	h.Data = uintptr(unsafe.Pointer(arr))
	h.Len = int(arrLen)
	h.Cap = int(arrLen)
	pulses := make([]Pulse, len(list)/2)
	// Convert original int array ([pulse, duration, pulse, duration, ...])
	// to Pulse struct array
	for i := 0; i < len(list)/2; i++ {
		var value byte = 0
		if list[i*2] != 0 {
			value = 1
		}
		pulses[i] = Pulse{Value: value,
			Duration: time.Duration(list[i*2+1]) * time.Microsecond}
	}
	return pulses, nil
}

// TODO write comment to function
func decodeByte(pulses []Pulse, start int) (byte, error) {
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
			return 0, fmt.Errorf("High edge value duration %v exceed "+
				"expected maximum amount %v", pulseH.Duration, HIGH_DUR_MAX)
		}
		if pulseH.Duration > HIGH_DUR_AVG {
			//fmt.Printf("bit %d is high\n", 7-i)
			b = b | (1 << uint(7-i))
		}
	}
	return byte(b), nil
}

// Decode bunch of pulse read from DHTxx sensors.
// Use pdf specifications from /docs folder to read 5 bytes and
// convert them to temperature and humidity.
func decodeDHTxxPulses(sensorType SensorType, pulses []Pulse) (temperature float32,
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
	// Decode 1st byte
	b0, err := decodeByte(pulses, 0)
	if err != nil {
		return -1, -1, err
	}
	// Decode 2nd byte
	b1, err := decodeByte(pulses, 16)
	if err != nil {
		return -1, -1, err
	}
	// Decode 3rd byte
	b2, err := decodeByte(pulses, 32)
	if err != nil {
		return -1, -1, err
	}
	// Decode 4th byte
	b3, err := decodeByte(pulses, 48)
	if err != nil {
		return -1, -1, err
	}
	// Decode 5th byte: control sum to verify all data received from sensor
	sum, err := decodeByte(pulses, 64)
	if err != nil {
		return -1, -1, err
	}
	// Produce data consistency check
	calcSum := byte(b0 + b1 + b2 + b3)
	if sum != calcSum {
		err := errors.New(spew.Sprintf(
			"CRCs doesn't match: checksum from sensor(%v) != "+
				"calculated checksum(%v=%v+%v+%v+%v)",
			sum, calcSum, b0, b1, b2, b3))
		return -1, -1, err
	} else {
		lg.Debugf("CRCs verified: checksum from sensor(%v) = calculated checksum(%v=%v+%v+%v+%v)",
			sum, calcSum, b0, b1, b2, b3)
	}
	// Debug output for 5 bytes
	lg.Debugf("Decoded from DHTxx sensor: [%d, %d, %d, %d, %d]", b0, b1, b2, b3, sum)
	// Extract temprature and humidity depending on sensor type
	temperature, humidity = 0.0, 0.0
	if sensorType == DHT11 {
		humidity = float32(b0)
		temperature = float32(b2)
	} else if sensorType == DHT22 {
		humidity = (float32(b0)*256 + float32(b1)) / 10.0
		temperature = (float32(b2&0x7F)*256 + float32(b3)) / 10.0
		if b2&0x80 != 0 {
			temperature *= -1.0
		}
	}
	// Additional check for data correctness
	if humidity > 100.0 {
		return -1, -1, fmt.Errorf("Humidity value exceed 100%%: %v", humidity)
	}
	// Success
	return temperature, humidity, nil
}

// Print bunch of pulses for debug purpose.
func printPulseArrayForDebug(pulses []Pulse) {
	// var buf bytes.Buffer
	// for i, pulse := range pulses {
	// 	buf.WriteString(fmt.Sprintf("pulse %3d: %v, %v\n", i,
	// 		pulse.Value, pulse.Duration))
	// }
	// lg.Debugf("Pulse count %d:\n%v", len(pulses), buf.String())
	lg.Debugf("Pulses received from DHTxx sensor: %v", pulses)
}

// Send activation request to DHTxx sensor via specific pin.
// Then decode pulses sent back with asynchronous
// protocol specific for DHTxx sensors.
//
// Input parameters:
// 1) sensor type: DHT11, DHT22 (aka AM2302);
// 2) pin number from GPIO connector to interract with sensor;
// 3) boost GPIO performance flag should be used for old devices
// such as Raspberry PI 1 (this will require root privileges).
//
// Return:
// 1) temperature in Celsius;
// 2) humidity in percent;
// 3) error if present.
func ReadDHTxx(sensorType SensorType, pin int,
	boostPerfFlag bool) (temperature float32, humidity float32, err error) {
	// Activate sensor and read data to pulses array
	pulses, err := dialDHTxxAndGetResponse(pin, boostPerfFlag)
	if err != nil {
		return -1, -1, err
	}
	// Output debug information
	printPulseArrayForDebug(pulses)
	// Decode pulses
	temp, hum, err := decodeDHTxxPulses(sensorType, pulses)
	if err != nil {
		return -1, -1, err
	}
	return temp, hum, nil
}

// Send activation request to DHTxx sensor via specific pin.
// Then decode pulses sent back with asynchronous
// protocol specific for DHTxx sensors. Retry n times in case of failure.
//
// Input parameters:
// 1) sensor type: DHT11, DHT22 (aka AM2302);
// 2) pin number from gadget GPIO to interract with sensor;
// 3) boost GPIO performance flag should be used for old devices
// such as Raspberry PI 1 (this will require root privileges);
// 4) how many times to retry until success either сounter is zeroed.
//
// Return:
// 1) temperature in Celsius;
// 2) humidity in percent;
// 3) number of extra retries data from sensor;
// 4) error if present.
func ReadDHTxxWithRetry(sensorType SensorType, pin int, boostPerfFlag bool,
	retry int) (temperature float32, humidity float32, retried int, err error) {
	ctx, cancel := context.WithCancel(context.Background())
	shell.CloseContextOnKillSignal(cancel)
	retried = 0
	for {
		temp, hum, err := ReadDHTxx(sensorType, pin, boostPerfFlag)
		if err != nil {
			if retry > 0 {
				lg.Warning(err)
				retry--
				retried++
				select {
				case <-ctx.Done():
					// Interrupt loop, if pending termination
					return -1, -1, retried, errors.New("Termination pending...")
				default:
					// Sleep before new attempt
					time.Sleep(1500 * time.Millisecond)
					continue
				}
			}
			return -1, -1, retried, err
		}
		return temp, hum, retried, nil
	}
}
