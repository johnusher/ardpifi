// mock_acc.go

package acc

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type mockACC struct {
	acc  chan<- ACCMessage
	done chan struct{}
	log  *log.Entry
}

func initMockACC(acc chan<- ACCMessage) (ACC, error) {
	return &mockACC{
		acc:  acc,
		done: make(chan struct{}, 0),
		log: log.WithFields(log.Fields{
			"mock": "acc",
		}),
	}, nil
}

func (m *mockACC) Run() error {
	m.log.Infof("Run()")

	// send a pseudo-random gps location every 2 seconds
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-m.done:
			m.log.Info("Exiting Run()")
			return nil
		case <-ticker.C:
			// temp := rand.int32()
			// temp := int8(27)
			bearing := rand.Float64() * 360
			roll := rand.Float64()
			tilt := rand.Float64()

			m.log.Infof("Sending %f %f %f", bearing, roll, tilt)

			m.acc <- ACCMessage{
				// Temp:    temp,
				Bearing: bearing,
				Roll:    roll,
				Tilt:    tilt,
			}
		}
	}
}

func (m *mockACC) Close() error {
	m.log.Infof("Close()")
	close(m.done)
	return nil
}
