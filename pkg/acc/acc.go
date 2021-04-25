package acc

// accelerometer, magnet etc sensor using the BNo055

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kpeu3i/bno055"
)

const (
	Pi = 3.14159265358979323846264338327950288419716939937510582097494459 // pi https://oeis.org/A000796
)

type ACC interface {
	Run() error
	Close() error
}

type ACCMessage struct {
	// Temp    int8
	Bearing float64
	Roll    float64
	Tilt    float64
}

type acc struct {
	acc    chan<- ACCMessage
	Sensor *bno055.Sensor
}

func Init(accChan chan<- ACCMessage, mock bool) (ACC, error) {
	if mock {
		return initMockACC(accChan)
	}

	return initACC(accChan)
}

func initACC(accChan chan<- ACCMessage) (ACC, error) {

	sensor, err := bno055.NewSensor(0x28, 3)
	if err != nil {
		panic(err)
	}

	err = sensor.UseExternalCrystal(true)
	if err != nil {
		panic(err)
	}

	status, err := sensor.Status()
	if err != nil {
		panic(err)
	}

	fmt.Printf("*** Status: system=%v, system_error=%v, self_test=%v\n", status.System, status.SystemError, status.SelfTest)

	_, err = sensor.AxisConfig()
	if err != nil {
		panic(err)
	}

	return &acc{
		accChan,
		sensor,
	}, nil

}

func (a *acc) Close() error {
	return a.Sensor.Close()
}

func (a *acc) Run() error {

	for {
		select {
		// case <-signals:
		// 	err := sensor.Close()
		// 	if err != nil {
		// 		panic(err)
		// 	}
		default:

			vector, err := a.Sensor.Euler()
			if err != nil {
				log.Errorf("acc error: %v", err)
			}

			bearing := float64(vector.X)
			roll := float64(vector.Y)
			tilt := float64(vector.Z)

			// acc, err := a.Sensor.LinearAccelerometer()
			// if err != nil {
			// 	log.Errorf("acc error: %v", err)
			// }

			// gyro, err := sensor.Gyroscope()
			// if err != nil {
			// 	log.Errorf("acc error: %v", err)
			// }

			// if err != nil {
			// 	log.Errorf("acc error: %v", err)
			// }

			// temp, err := a.Sensor.Temperature()
			// if err != nil {
			// 	panic(err)
			// }

			a.acc <- ACCMessage{
				// Temp:    temp,
				Bearing: bearing,
				Roll:    roll,
				Tilt:    tilt,
			}

		}

		time.Sleep(150 * time.Millisecond)
	}

}
