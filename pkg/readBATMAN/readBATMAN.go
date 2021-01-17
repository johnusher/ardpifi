package readBATMAN

import (
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/johnusher/ardpifi/pkg/iface"
	log "github.com/sirupsen/logrus"
)

type ReadBATMAN struct {
	messages chan<- uint32
	FarEndIP *net.IP
	Conn     *net.UDPConn
}

const (
	port      = 4200
	msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc
)

func Init(messages chan<- uint32, noHardware bool) (*ReadBATMAN, error) {
	// err := termbox.Init()
	// if err != nil {
	// 	return nil, err
	// }

	myIP := net.IP{}
	// myPings := uint32(0)

	i, err := iface.InterfaceByName(ifaceName, noHardware)
	if err != nil {
		log.Errorf("InterfaceByName failed: %s", err)
		return nil, err
	}

	addrs, err := i.Addrs()
	if err != nil {
		log.Errorf("Failed to get addresses for interface %+v: %s", i, err)
		return nil, err
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
		log.Error(err)
		return nil, err
	}

	return &ReadBATMAN{
		messages,
		nil,
		conn,
	}, nil
}

func (k *ReadBATMAN) Run() error {
	defer func() {
		close(k.messages)
	}()

	log.Info("LEDMesh starting up")

	log.Infof("Listening as %+v", k.Conn.LocalAddr().(*net.UDPAddr))

	buffIn := make([]byte, msgSize) // received via BATMAM
	// buffOut := make([]byte, msgSize) // sent to batman
	// copy(buffOut[0:4], myIP)

	// bcast := &net.UDPAddr{Port: port, IP: net.IPv4(172, 27, 255, 255)}
	// pingAt := time.Now()

	for {

		if n, addr, err := k.Conn.ReadFromUDP(buffIn); err == nil {

			if n == msgSize {
				pings := uint32(buffIn[4]) +
					uint32(buffIn[5])<<8 +
					uint32(buffIn[6])<<16 +
					uint32(buffIn[7])<<24
					// 4 bytes

				log.Infof("%+v: %s: %d", addr, net.IP(buffIn[0:4]), pings)

				k.messages <- pings // send to output
				// k.FarEndIP = net.IP(buffIn[0:4])
			} else {
				log.Errorf("Received unexpected message length from %+v: %d", addr, n)
			}
		} else if ne, ok := err.(*net.OpError); !ok || !ne.Timeout() {
			log.Errorf("ReadFromUDP failed with %s", err)
		}

	}
}
