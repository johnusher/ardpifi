// mock_acc.go

package acc

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type mockACC struct {
	acc chan<- ACCMessage
	// acc2 chan<- ACCMessage2
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
	ticker := time.NewTicker(20 * time.Millisecond)

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

			quat_w := rand.Float64()
			quat_x := rand.Float64()
			quat_y := rand.Float64()
			quat_z := rand.Float64()

			// m.log.Infof("Sending %f %f %f", bearing, roll, tilt)

			// m.acc <- ACCMessage{
			// 	// Temp:    temp,
			// 	Bearing: bearing,
			// 	Roll:    roll,
			// 	QuatW:   quat_w,
			// 	QuatX:   quat_x,
			// 	QuatY:   quat_y,
			// 	QuatZ:   quat_z,
			// }

			m.acc <- ACCMessage{
				// Temp:    temp,
				Bearing: bearing,
				Roll:    roll,
				Tilt:    tilt,
				QuatW:   quat_w,
				QuatX:   quat_x,
				QuatY:   quat_y,
				QuatZ:   quat_z,
			}

			// m.acc2 <- ACCMessage2{
			// 	// Temp:    temp,
			// 	// Bearing: bearing,
			// 	// Roll:    roll,
			// 	// Tilt   tilt,
			// 	QuatW: quat_w,
			// 	QuatX: quat_x,
			// 	QuatY: quat_y,
			// 	QuatZ: quat_z,
			// }
		}
	}
}

func (m *mockACC) Close() error {
	m.log.Infof("Close()")
	close(m.done)
	return nil
}
