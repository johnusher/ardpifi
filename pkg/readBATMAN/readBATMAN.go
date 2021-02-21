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
	messages chan<- []byte
	FarEndIP *net.IP
	Conn     *net.UDPConn
	log      *log.Entry
}

const (
	port = 4200
	// msgSize   = net.IPv4len + 4 // IP + uint32
	msgSize   = 1000 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc
)

func Init(messages chan<- []byte, mock bool, bcastIP net.IP) (*ReadBATMAN, error) {
	// err := termbox.Init()
	// if err != nil {
	// 	return nil, err
	// }

	myIP := net.IP{}
	// myPings := uint32(0)

	i, err := iface.InterfaceByName(ifaceName, mock, bcastIP)
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
		if ip4 != nil && ip4[0] == bcastIP.To4()[0] {
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
		log.WithFields(log.Fields{
			"package": "readBATMAN",
			"bcastIP": bcastIP,
		}),
	}, nil
}

func (k *ReadBATMAN) Run() error {
	defer func() {
		close(k.messages)
	}()

	k.log.Info("LEDMesh starting up")

	k.log.Infof("Listening as %+v", k.Conn.LocalAddr().(*net.UDPAddr))

	buffIn := make([]byte, msgSize) // received via BATMAM
	// buffOut := make([]byte, msgSize) // sent to batman
	// copy(buffOut[0:4], myIP)

	// bcast := &net.UDPAddr{Port: port, IP: net.IPv4(172, 27, 255, 255)}
	// pingAt := time.Now()

	for {
		n, addr, err := k.Conn.ReadFromUDP(buffIn)
		if err != nil {
			k.log.Errorf("ReadFromUDP failed with %s", err)
			continue
		}

		if n > msgSize {
			k.log.Errorf("Received unexpected message length from %+v: %d", addr, n)
			continue
		}

		// pings := uint32(buffIn[4]) +
		// 	uint32(buffIn[5])<<8 +
		// 	uint32(buffIn[6])<<16 +
		// 	uint32(buffIn[7])<<24
		// 	// 4 bytes
		// msg := buffIn[0:n]

		// k.log.Infof("received mesh-age")

		buffOut := make([]byte, n)
		copied := copy(buffOut, buffIn)
		if copied != n {
			k.log.Errorf("Copied wrong number of bytes, expected %d, got %d, for message: %s", n, copied, buffIn)
			continue
		}

		k.messages <- buffOut // send to output
		// k.FarEndIP = net.IP(buffIn[0:4])
	}
}
