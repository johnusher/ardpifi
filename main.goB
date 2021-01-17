package main

import (
	"net"
	"os"
	"os/signal"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	port      = 4200
	msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc
)

func main() {
	log.Info("LEDMesh starting up")

	myIP := net.IP{}
	myPings := uint32(0)

	i, err := net.InterfaceByName(ifaceName)
	if err != nil {
		log.Fatalf("InterfaceByName failed: %s", err)
	}

	addrs, err := i.Addrs()
	if err != nil {
		log.Fatalf("Failed to get addresses for interface %+v: %s", i, err)
	}

	for _, addr := range addrs {
		ipnet := addr.(*net.IPNet)
		ip4 := ipnet.IP.To4()
		if ip4 != nil && ip4[0] == 172 {
			myIP = ip4
		}
	}

	log.Infof("Serving at %s", myIP)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Listening as %+v", conn.LocalAddr().(*net.UDPAddr))

	buffIn := make([]byte, msgSize)
	buffOut := make([]byte, msgSize)
	copy(buffOut[0:4], myIP)

	bcast := &net.UDPAddr{Port: port, IP: net.IPv4(172, 27, 255, 255)}
	pingAt := time.Now()

	for {
		if err := conn.SetReadDeadline(pingAt); err != nil {
			log.Fatal(err)
		}

		// read
		if n, addr, err := conn.ReadFromUDP(buffIn); err == nil {
			if n == msgSize {
				pings := uint32(buffIn[4]) +
					uint32(buffIn[5])<<8 +
					uint32(buffIn[6])<<16 +
					uint32(buffIn[7])<<24

				log.Infof("%+v: %s: %d", addr, net.IP(buffIn[0:4]), pings)
			} else {
				log.Errorf("Received unexpected message length from %+v: %d", addr, n)
			}
		} else if ne, ok := err.(*net.OpError); !ok || !ne.Timeout() {
			log.Errorf("ReadFromUDP failed with %s", err)
		}

		// write
		if time.Now().After(pingAt) {
			buffOut[4] = byte(myPings & 0x000000ff)
			buffOut[5] = byte(myPings & 0x0000ff00 >> 8)
			buffOut[6] = byte(myPings & 0x00ff0000 >> 16)
			buffOut[7] = byte(myPings & 0xff000000 >> 24)
			if _, err := conn.WriteToUDP(buffOut, bcast); err != nil {
				log.Fatal(err)
			}
			pingAt = time.Now().Add(interval)
			myPings++
		}

		select {
		case <-sig:
			log.Info("Interrupt signal received, exiting")
			return
		default:
		}
	}
}
