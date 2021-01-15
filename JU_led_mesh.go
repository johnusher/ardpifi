// JU_led_mesh

// based on https://github.com/siggy/ledmesh/blob/master/main.go

package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/johnusher/ardpifi/pkg/keyboard"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

const (
	port      = 4200
	msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc
)

func main() {

	// Find the device that represents the arduino serial
	// connection. NB this is kinda janky- we should have a system to robustly detect a duino,
	// eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	c := &serial.Config{Name: findArduino(), Baud: 9600, ReadTimeout: time.Second * 1}

	s, err := serial.OpenPort(c)
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
	log.Print("%q", buf[:n])

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

	// go kb.Run()

	buf = make([]byte, 5)

	//  now setup BATMAN:

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

	buffIn := make([]byte, msgSize)  // received via BATMAM
	buffOut := make([]byte, msgSize) // sent to batman
	copy(buffOut[0:4], myIP)

	bcast := &net.UDPAddr{Port: port, IP: net.IPv4(172, 27, 255, 255)}
	pingAt := time.Now()

	// run kb and BATMAN:

	go kb.Run()

	for {

		if err := conn.SetReadDeadline(pingAt); err != nil {
			log.Fatal(err)
		}

		// read: NB i want to change this so we get an interupt when there is a UDP message!
		if n, addr, err := conn.ReadFromUDP(buffIn); err == nil {
			if n == msgSize {
				pings := uint32(buffIn[4]) +
					uint32(buffIn[5])<<8 +
					uint32(buffIn[6])<<16 +
					uint32(buffIn[7])<<24
					// 4 bytes

				log.Infof("%+v: %s: %d", addr, net.IP(buffIn[0:4]), pings)
			} else {
				log.Errorf("Received unexpected message length from %+v: %d", addr, n)
			}
		} else if ne, ok := err.(*net.OpError); !ok || !ne.Timeout() {
			log.Errorf("ReadFromUDP failed with %s", err)
		}

		// listen on the keys channel for key presses
		select {
		case key, more := <-keys:
			if !more {
				log.Infof("keyboard listener closed\n")

				// termbox closed, block until ctrl-c is called
				<-stop

				log.Infof("exiting")
				return
			}
			log.Infof("key pressed: %s / %d / 0x%X / 0%o \n", string(key), key, key, key)

			// _, err := s.Write([]byte(0))

			n, err = s.Write([]byte(string(key)))
			if err != nil {
				log.Errorf("2. failed to write to serial port: %s", err)
				return
			}

			n, err = s.Read(buf)
			if err != nil {
				log.Errorf("serial port read error, %s", err)
			}
			log.Infof("serial return %s / %d / 0x%X / 0%o \n", string(buf[:n]), buf[:n], buf[:n], buf[:n])
			// log.Infof("%q", buf[:n])

			// now send the key over BATMAN:

			// buf := make([]byte, 1)
			// _ = utf8.EncodeRune(buf, key)

			// myPings = buf

			myPings = uint32(key) // convert rune to uint32
			// write
			// if time.Now().After(pingAt) {
			buffOut[4] = byte(myPings & 0x000000ff)
			buffOut[5] = byte(myPings & 0x0000ff00 >> 8)
			buffOut[6] = byte(myPings & 0x00ff0000 >> 16)
			buffOut[7] = byte(myPings & 0xff000000 >> 24)
			if _, err := conn.WriteToUDP(buffOut, bcast); err != nil {
				log.Fatal(err)
			}
			pingAt = time.Now().Add(interval)
			// myPings++
			// }
		default:
			// fall through, add a sleep here if you want to slow things down
		}
	}

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
