// JU_led_mesh

// based on https://github.com/siggy/ledmesh/blob/master/main.go

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johnusher/ardpifi/pkg/iface"
	"github.com/johnusher/ardpifi/pkg/keyboard"
	"github.com/johnusher/ardpifi/pkg/port"
	"github.com/johnusher/ardpifi/pkg/readBATMAN"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

const (
	batPort   = 4200
	msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc

	batBcast   = "172.27.255.255"
	localBcast = "127.0.0.1"
)

func main() {
	noHardware := flag.Bool("no-hardware", false, "run without hardware dependencies")
	flag.Parse()

	bcastIP := net.ParseIP(batBcast)
	if *noHardware {
		bcastIP = net.ParseIP(localBcast)
	}

	// Find the device that represents the arduino serial
	// connection. NB this is kinda janky- we should have a system to robustly detect a duino,
	// eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	c := &serial.Config{Name: findArduino(), Baud: 9600, ReadTimeout: time.Second * 1}

	s, err := port.OpenPort(c, *noHardware)
	if err != nil {
		log.Errorf("OpenPort error: %s", err)
		return
	}

	// When connecting to an older revision Arduino, you need to wait
	// a little while it resets.
	time.Sleep(1 * time.Second)

	n, err := s.Write([]byte("C"))
	// send a C for Connect signal to the board and check response
	if err != nil {
		log.Errorf("failed to write to port: %s", err)
		return
	}

	// read return message from duino:
	buf := make([]byte, 1)
	n, err = s.Read(buf)
	if err != nil {
		log.Errorf("serial port read error, %s", err)
	}
	log.Info("%q", buf[:n])

	// now check if got the correct response:

	// Setup remote (terminal) KB:
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	keys := make(chan rune)

	kb, err := keyboard.Init(keys)
	if err != nil {
		log.Errorf("failed to initialize keyboard: %s", err)
		return
	}

	//  now setup BATMAN:

	// log.Info("LEDMesh starting up")

	myIP := net.IP{}

	i, err := iface.InterfaceByName(ifaceName, *noHardware, bcastIP)
	if err != nil {
		log.Errorf("InterfaceByName failed: %s", err)
		return
	}

	addrs, err := i.Addrs()
	if err != nil {
		log.Errorf("Failed to get addresses for interface %+v: %s", i, err)
		return
	}

	for _, addr := range addrs {
		ipnet := addr.(*net.IPNet)
		ip4 := ipnet.IP.To4()
		if ip4 != nil && ip4[0] == bcastIP.To4()[0] {
			myIP = ip4
		}
	}

	log.Infof("Serving at %s", myIP)

	// sig := make(chan os.Signal, 1)
	// signal.Notify(sig, os.Interrupt, os.Kill)

	// conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: port})
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// pingAt := time.Now()

	// init BATMAN:
	messages := make(chan []byte)
	bm, err := readBATMAN.Init(messages, *noHardware, bcastIP)
	if err != nil {
		log.Errorf("failed to initialize readBATMAN: %s", err)
		return
	}

	// run kb and BATMAN:

	go kb.Run()
	go bm.Run()

	errs := make(chan error)

	go func() {
		errs <- messageLoop(messages, s, myIP)
	}()
	go func() {
		errs <- keyLoop(keys, s, myIP, bcastIP, bm)
	}()

	// block until ctrl-c or one of the loops returns an error
	select {
	case <-stop:
	case <-errs:
	}
}

func messageLoop(messages <-chan []byte, s port.Port, myIP net.IP) error {
	log.Info("Starting message loop")

	for {
		// listen on the keys channel for key presses AND listen for new BATMAN message
		message, _ := <-messages

		if len(message) != msgSize {
			log.Errorf("Received unexpected message length %d, expected %d: %x", len(message), msgSize, message)
			continue
		}

		ip := net.IP(message[0:4])
		pings := uint32(message[4]) +
			uint32(message[5])<<8 +
			uint32(message[6])<<16 +
			uint32(message[7])<<24

		if myIP.Equal(net.IP(message[0:4])) {
			log.Infof("received message from my own IP: %s / %s", ip, string(pings))
		} else {
			log.Infof("received message from other IP: %s / %s", ip, string(pings))

			// write to duino:
			_, err := s.Write([]byte(string(message)))
			if err != nil {
				log.Errorf("3. failed to write to serial port: %s", err)
				return err
			}

		}

		log.Infof("BATMAN message : %s / %d / 0x%X / 0%o \n", string(pings), pings, pings, pings)

	}
}

func keyLoop(keys <-chan rune, s port.Port, myIP net.IP, bcastIP net.IP, bm *readBATMAN.ReadBATMAN) error {
	log.Info("Starting key loop")

	buf := make([]byte, 5)

	buffOut := make([]byte, msgSize) // sent to batman
	copy(buffOut[0:4], myIP)

	bcast := &net.UDPAddr{Port: batPort, IP: bcastIP}

	for {
		key, more := <-keys
		if !more {
			log.Infof("keyboard listener closed\n")
			// termbox closed, block until ctrl-c is called
			log.Infof("exiting")
			return nil
		}
		log.Infof("key pressed: %s / %d / 0x%X / 0%o \n", string(key), key, key, key)

		// now send the key over BATMAN:
		// buf := make([]byte, 1)
		// _ = utf8.EncodeRune(buf, key)
		myPings := uint32(key) // convert rune to uint32
		// write
		// if time.Now().After(pingAt) {
		buffOut[4] = byte(myPings & 0x000000ff)
		buffOut[5] = byte(myPings & 0x0000ff00 >> 8)
		buffOut[6] = byte(myPings & 0x00ff0000 >> 16)
		buffOut[7] = byte(myPings & 0xff000000 >> 24)
		if _, err := bm.Conn.WriteToUDP(buffOut, bcast); err != nil {
			log.Error(err)
			return err
		}
		// pingAt = time.Now().Add(interval)
		myPings++

		// write to duino: NB maybe insert a wait before here so all pi's send the new duino command at a similar time
		n, err := s.Write([]byte(string(key)))
		if err != nil {
			log.Errorf("2. failed to write to serial port: %s", err)
			return err
		}
		// read response from duin (not necessary)
		n, err = s.Read(buf)
		if err != nil {
			log.Errorf("serial port read error, %s", err)
		}
		log.Infof("serial return %s / %d / 0x%X / 0%o \n", string(buf[:n]), buf[:n], buf[:n], buf[:n])
		// }

	}

	return nil
}

// findArduino looks for the file that represents the Arduino
// serial connection. Returns the fully qualified path to the
// device if we are able to find a likely candidate for an
// Arduino, otherwise an empty string if unable to find
// something that 'looks' like an Arduino device.
func findArduino() string {
	contents, _ := ioutil.ReadDir("/dev")

	// Look for what is mostly likely the Arduino device
	// NB this is kinda janky- we should have a system to robustly detect a duino, eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	// JU: on my RASPI the legit Aurdion Uno shows in ttyACM0, but my fake nano +CH340-Chip shows on ttyUSB0
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyUSB") || strings.Contains(f.Name(), "ttyACM0") {
			fmt.Println("Duino found at /dev/", f.Name())
			return "/dev/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}

func findUSBKeyboard() string {
	contents, _ := ioutil.ReadDir("/dev/input")

	// Look for what is mostly likely the local USB KB

	for _, f := range contents {
		if strings.Contains(f.Name(), "event") {
			fmt.Println("USB KB found at /dev/input/", f.Name())
			return "/dev/input/" + f.Name()
		}
	}

	// Have not been able to find a USB device that 'looks'
	// like an Arduino.
	return ""
}
