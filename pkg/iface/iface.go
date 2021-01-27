package iface

import (
	"net"

	log "github.com/sirupsen/logrus"
)

type Interface interface {
	Addrs() ([]net.Addr, error)
}

type mockInterface struct {
	ip  net.IP
	log *log.Entry
}

func InterfaceByName(name string, mock bool, mockBcastIP net.IP) (Interface, error) {
	if mock {
		return &mockInterface{
			ip: mockBcastIP,
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

	return []net.Addr{&net.IPNet{IP: m.ip}}, nil
}
