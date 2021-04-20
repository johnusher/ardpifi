package gps

import (
	"math/rand"
	"time"

	log "github.com/sirupsen/logrus"
)

type mockGPS struct {
	gps  chan<- GPSMessage
	done chan struct{}
	log  *log.Entry
}

func initMockGPS(gps chan<- GPSMessage) (GPS, error) {
	return &mockGPS{
		gps:  gps,
		done: make(chan struct{}, 0),
		log: log.WithFields(log.Fields{
			"mock": "gps",
		}),
	}, nil
}

func (m *mockGPS) Run() error {
	m.log.Infof("Run()")

	// send a pseudo-random gps location every 2 seconds
	ticker := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-m.done:
			m.log.Info("Exiting Run()")
			return nil
		case <-ticker.C:
			lat := 52.534634 + rand.Float64()*0.0001
			long := 13.347364 + rand.Float64()*0.0001
			// fixQuality := uint16(1)
			hdop := 0.9 + rand.Float64()*0.1

			m.log.Infof("Sending %f %f %f", lat, long, hdop)

			m.gps <- GPSMessage{
				Lat:  lat,
				Long: long,
				// FixQuality: fixQuality,
				HDOP: hdop,
			}
		}
	}
}

func (m *mockGPS) Close() error {
	m.log.Infof("Close()")
	close(m.done)
	return nil
}
