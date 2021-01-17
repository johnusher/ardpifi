package iface

import (
	"net"

	log "github.com/sirupsen/logrus"
)

type Interface interface {
	Addrs() ([]net.Addr, error)
}

type mockInterface struct {
	log *log.Entry
}

func InterfaceByName(name string, noHardware bool) (Interface, error) {
	if noHardware {
		return &mockInterface{
			log: log.WithFields(log.Fields{
				"mock": "interface",
				"name": name,
			}),
		}, nil
	}

	return net.InterfaceByName(name)
}

func (m *mockInterface) Addrs() ([]net.Addr, error) {
	log.Info("Addrs")

	return []net.Addr{
		&net.IPNet{
			IP: net.ParseIP("172.27.42.42"),
		},
	}, nil
}
