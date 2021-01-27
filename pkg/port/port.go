package port

import (
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

type Port interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Flush() error
	Close() error
}

type mockPort struct {
	b          []byte
	sync.Mutex // protects byte buffer
	log        *log.Entry
}

func OpenPort(c *serial.Config, mock bool) (Port, error) {
	if mock {
		return &mockPort{
			log: log.WithFields(log.Fields{
				"mock":   "port",
				"config": c,
			}),
		}, nil
	}
	return serial.OpenPort(c)
}

func (m *mockPort) Read(b []byte) (int, error) {
	m.log.Infof("Read: %s", m.b)

	m.Lock()
	defer m.Unlock()

	return copy(b, m.b), nil
}
func (m *mockPort) Write(b []byte) (int, error) {
	m.log.Infof("Write: %s", b)

	m.Lock()
	defer m.Unlock()

	m.b = make([]byte, len(b))
	return copy(m.b, b), nil
}
func (m *mockPort) Flush() error {
	m.log.Info("Flush")
	return nil
}
func (m *mockPort) Close() error {
	m.log.Info("Close")
	return nil
}
