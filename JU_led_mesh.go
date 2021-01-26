// JU_led_mesh

// based on https://github.com/siggy/ledmesh/blob/master/main.go

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// "github.com/qinxin0720/lcd1602"

	device "github.com/d2r2/go-hd44780"
	i2c "github.com/d2r2/go-i2c"
	"github.com/johnusher/ardpifi/pkg/gps"
	"github.com/johnusher/ardpifi/pkg/iface"
	"github.com/johnusher/ardpifi/pkg/keyboard"
	"github.com/johnusher/ardpifi/pkg/lcd"
	"github.com/johnusher/ardpifi/pkg/port"
	"github.com/johnusher/ardpifi/pkg/readBATMAN"
	"github.com/johnusher/ardpifi/pkg/web"
	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"
)

const (
	batPort = 4200
	// msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc

	batBcast   = "172.27.255.255"
	localBcast = "127.0.0.1"
)

type ChatRequest struct {
	Latf  float64
	Longf float64
	ID    string
	Key   rune
}

func main() {
	raspID := flag.String("rasp-id", "raspi 1", "unique raspberry pi ID")
	webAddr := flag.String("web-addr", ":8080", "address to serve web on")
	noHardware := flag.Bool("no-hardware", false, "run without hardware dependencies")
	noLCD := flag.Bool("no-lcd", false, "run without lcd display")
	logLevel := flag.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")

	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("failed to parse log level [%s]: %s", *logLevel, err)
		return
	}
	log.SetLevel(level)

	// // LCD:

	var i2cDevice *i2c.I2C
	if !*noLCD {
		var err error
		i2cDevice, err = i2c.NewI2C(0x27, 1)
		if err != nil {
			log.Errorf("error opening LCD on I2C, %s", err)
			return
		}
		defer i2cDevice.Close()
	}

	lcd, err := lcd.New(i2cDevice, *noLCD)
	if err != nil {
		log.Errorf("errormaking LCD device, %s", err)
		return
	}

	lcd.BacklightOn()
	lcd.Clear()
	// lcd.SetStrobeDelays(300)   // TODO: doenst work but should!

	web := web.InitWeb(*webAddr)
	log.Infof("web: %+v", web)

	bcastIP := net.ParseIP(batBcast)
	if *noHardware {
		bcastIP = net.ParseIP(localBcast)
	}

	// Find the device that represents the arduino serial
	// connection. NB this is kinda janky- we should have a system to robustly detect a duino,
	// eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	// c := &serial.Config{Name: findArduino(), Baud: 9600, ReadTimeout: time.Second * 1}
	// c := &serial.Config{Name: findArduino(), Baud: 19200, ReadTimeout: time.Second * 1}
	c := &serial.Config{Name: findArduino(), Baud: 115200, ReadTimeout: time.Second * 1}

	s, err := port.OpenPort(c, *noHardware)
	if err != nil {
		log.Errorf("OpenPort error: %s", err)
		return
	}

	// When connecting to an older revision Arduino, you need to wait
	time.Sleep(1 * time.Second)

	// Setup keyboard input:
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	keys := make(chan rune)

	kb, err := keyboard.Init(keys)
	if err != nil {
		log.Errorf("failed to initialize keyboard: %s", err)
		return
	}

	//  now setup BATMAN:

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

	// write to LCD:
	lcd.SetPosition(0, 0)
	_ = lcd.ShowMessage("Starting", device.SHOW_LINE_1)

	// lcd.SetPosition(1, 0)

	// _ = lcd.ShowMessage(string(myIP), device.SHOW_LINE_2)

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

	gpsChan := make(chan gps.GPSMessage)
	g, err := gps.Init(gpsChan, *noHardware)
	if err != nil {
		log.Errorf("failed to initialize gps: %s", err)
		return
	}
	defer g.Close()

	// run kb and BATMAN:
	go kb.Run()
	go bm.Run()
	go g.Run()

	errs := make(chan error)

	go func() {
		errs <- messageLoop(messages, s, *raspID, lcd, web)
	}()
	go func() {
		errs <- broadcastLoop(keys, gpsChan, s, *raspID, bcastIP, bm)
	}()
	go func() {
		// handle key presses from web, send to messages channel
		for {
			phoneEvent, more := <-web.Phone()
			if !more {
				log.Errorf("web phoneEvent channel closed")
				return
			}

			if len(phoneEvent.Key) == 0 {
				continue
			}

			keys <- []rune(phoneEvent.Key)[0]
		}
	}()

	// block until ctrl-c or one of the loops returns an error
	select {
	case <-stop:
	case <-errs:
	}
}

func messageLoop(messages <-chan []byte, s port.Port, raspID string, lcd lcd.LCD, web *web.Web) error {
	log.Info("Starting message loop")

	for {
		// listen on the keys channel for key presses AND listen for new BATMAN message
		message, _ := <-messages

		jsonMessage := ChatRequest{}

		// make json:
		err := json.Unmarshal(message, &jsonMessage)
		if err != nil {
			log.Errorf("Unmarshal failed: %s", err)
			return err
		}
		// ip := net.IP(message[0:4])
		// pings := uint32(message[4]) +
		// 	uint32(message[5])<<8 +
		// 	uint32(message[6])<<16 +
		// 	uint32(message[7])<<24

		if jsonMessage.ID == raspID {
			// message from self:
			// msg := fmt.Sprintf("received message from self: %+v", jsonMessage)
			// log.Info(msg)
			// web.Render(msg)
		} else {

			//  message from other:
			msg := fmt.Sprintf("received message from other: %+v", jsonMessage)
			log.Info(msg)
			web.Render(msg)

			if string(jsonMessage.Key) != "x" {

				log.Infof("key from other %s \n", (string(jsonMessage.Key))) // this doesnt point to the "Key" element of the struct!

				// msg = fmt.Sprintf("received message from other raspi: %s", jsonMessage)
				// log.Info(msg)
				// web.Render(msg)

				// write to duino:
				s.Flush()
				_, err := s.Write([]byte(string(jsonMessage.Key)))
				// message[4] gives me the letter "t", perhaps as message = {Latf:52.534587 Longf:13.347233 ID:raspi 1 Key:0}

				if err != nil {
					log.Errorf("3. failed to write to serial port: %s", err)
					//return err
				}
				s.Flush()

			}

			// // write to LCD:
			// lcd.Clear()
			// lcd.SetPosition(0, 0)
			// // fmt.Fprint(lcd, t.Format("Message received:"))
			// _ = lcd.ShowMessage("Message received:", device.SHOW_LINE_1)
			// lcd.SetPosition(1, 0)
			// // fmt.Fprint(lcd, t.Format(string(message[4])))
			// _ = lcd.ShowMessage(string(message[4]), device.SHOW_LINE_2)

		}

		// log.Infof("BATMAN message : %s / %d / 0x%X / 0%o \n", string(pings), pings, pings, pings)

	}
}

func broadcastLoop(keys <-chan rune, gps <-chan gps.GPSMessage, s port.Port, raspID string, bcastIP net.IP, bm *readBATMAN.ReadBATMAN) error {
	log.Info("Starting key loop")

	// buf := make([]byte, 5)   // this was used for serial return from duino

	// buffOut := make([]byte, msgSize) // sent to mesh
	// copy(buffOut[0:4], myIP)

	bcast := &net.UDPAddr{Port: batPort, IP: bcastIP}

	for {
		select {

		case gpsMessage, more := <-gps:
			if !more {
				log.Infof("gps channel closed\n")
				log.Infof("exiting")
				return nil
			}

			// log.Infof("Local GPS Message received: %+v", gpsMessage)

			// make struct we send over udp:
			initChatRequest := ChatRequest{
				Latf:  gpsMessage.Lat,
				Longf: gpsMessage.Long,
				ID:    raspID,
				Key:   'x',
			}

			// make json:
			jsonRequest, err := json.Marshal(initChatRequest)
			if err != nil {
				log.Errorf("Marshal Register information failed: %s", err)
				return err
			}
			_, err = bm.Conn.WriteToUDP(jsonRequest, bcast)
			if err != nil {
				log.Error(err)
				return err
			}

			// if _, err := bm.Conn.WriteToUDP(gpsMessage, bcast); err != nil {
			// 	log.Error(err)
			// 	return err
			// }

		case key, more := <-keys:
			if !more {
				log.Infof("keyboard listener closed\n")
				// termbox closed, block until ctrl-c is called
				log.Infof("exiting")
				return nil
			}
			log.Infof("key pressed: %s / %d / 0x%X / 0%o \n", string(key), key, key, key)

			initChatRequest := ChatRequest{
				ID:  raspID,
				Key: key,
			}

			// make json:
			jsonRequest, err := json.Marshal(initChatRequest)
			if err != nil {
				log.Errorf("Marshal Register information failed: %s", err)
				return err
			}
			_, err = bm.Conn.WriteToUDP(jsonRequest, bcast)
			if err != nil {
				log.Error(err)
				return err
			}

			// // now send the key over BATMAN:
			// // buf := make([]byte, 1)
			// // _ = utf8.EncodeRune(buf, key)
			// myPings := uint32(key) // convert rune to uint32
			// // write
			// // if time.Now().After(pingAt) {
			// buffOut[4] = byte(myPings & 0x000000ff)
			// buffOut[5] = byte(myPings & 0x0000ff00 >> 8)
			// buffOut[6] = byte(myPings & 0x00ff0000 >> 16)
			// buffOut[7] = byte(myPings & 0xff000000 >> 24)
			// if _, err := bm.Conn.WriteToUDP(buffOut, bcast); err != nil {
			// 	log.Error(err)
			// 	return err
			// }
			// // pingAt = time.Now().Add(interval)
			// myPings++

			// write to duino: NB maybe insert a wait before here so all pi's send the new duino command at a similar time
			_, err = s.Write([]byte(string(key)))
			if err != nil {
				log.Errorf("2. failed to write to serial port: %s", err)
				return err
			}

			// // read response from duin (not necessary)
			// n, err = s.Read(buf)
			// if err != nil {
			// 	log.Errorf("serial port read error, %s", err)
			// }
			// log.Infof("serial return %s / %d / 0x%X / 0%o \n", string(buf[:n]), buf[:n], buf[:n], buf[:n])
			// // }

		}
	}
}

// findArduino looks for the file that represents the Arduino
// serial connection.

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
