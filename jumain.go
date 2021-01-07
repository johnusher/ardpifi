// public ardpifi

// to check which I2C device are connected run i2cdetect -y 1 (apt install i2c-tools )

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"

	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"

	"github.com/johnusher/ardpifi/pkg/keyboard"
	// "github.com/johnusher/ardpifi/pkg/localkeyboard"
)

func main() {

	// // check if we have USB keyboard attached:
	// localKB, localKB_det := os.Open(findUSBKeyboard())
	// if localKB_det != nil {
	// 	log.Print("No local USB keyboard found")
	// } else {
	// 	defer localKB.Close() // do we need to do this?
	// }

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

	// localkeys := make(chan rune)
	// localkb, err := localkeyboard.Init(localkeys)

	// if err != nil {
	// 	log.Errorf("failed to initialize local keyboard: %s", err)
	// 	return
	// }

	go kb.Run()

	// go localkb.Run()

	// x := <-localkeys

	// log.Infof("jj", string(x))
	buf = make([]byte, 128)
	for {
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

		}
	}

	// everything below here is unreachable due to the event loop above that exits

	defer logger.FinalizeLogger()
	fmt.Println("started i2c'ing\n")

	// Create new connection to Arduino:
	// I2C bus on line 1 with address 0x08
	i2c, err := i2c.NewI2C(0x08, 1)
	if err != nil {
		log.Errorf("failed to initialize i2c: %s", err)
		return
	}
	// Free I2C connection on exit
	defer i2c.Close()

	// Uncomment/comment next line to suppress/increase verbosity of output
	err = logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	if err != nil {
		log.Errorf("failed to ChangePackageLogLevel: %s", err)
		return
	}

	// // this next bit does not work:!!
	// from https://github.com/d2r2/go-i2c
	// ....
	// // Here goes code specific for sending and reading data
	// // to and from device connected via I2C bus, like:
	// _, err := i2c.Write([]byte{0x1, 0xF3})
	// if err != nil { log.Fatal(err) }
	// ....

	// write data to I2C
	n, err = i2c.WriteBytes([]byte{0x1, 0xF3})
	if err != nil {
		log.Errorf("failed to WriteBytes: %s", err)
		return
	}

	log.Printf("wrote %d bytes", n)

	// Uncomment/comment next line to suppress/increase verbosity of output
	err = logger.ChangePackageLogLevel("i2c", logger.InfoLevel)
	if err != nil {
		log.Errorf("failed to ChangePackageLogLevel: %s", err)
		return
	}

	fmt.Println("ended")
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

	// JU: on my RASPI it shows in ttyAMA0
	for _, f := range contents {
		if strings.Contains(f.Name(), "ttyACM0") {
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

	// Look for what is mostly likely the Arduino device
	// NB this is kinda janky- we should have a system to robustly detect a duino, eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	// JU: on my RASPI it shows in ttyAMA0
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

// sendArduinoCommand transmits a new command over the nominated serial
// port to the arduino. Returns an error on failure. Each command is
// identified by a single byte and may take one argument (a float).
// func sendArduinoCommand(
// 	command byte, serialPort io.ReadWriteCloser) error {
// 	if serialPort == nil {
// 		return nil
// 	}

// 	// // Package argument for transmission
// 	// bufOut := new(bytes.Buffer)
// 	// err := binary.Write(bufOut, binary.LittleEndian, argument)
// 	// if err != nil {
// 	// 	return err
// 	// }

// 	// Transmit command and argument down the pipe.
// 	for _, v := range []byte{command} {
// 		_, err = serialPort.Write(v)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func sendArduinoCommand(
// 	command byte, argument float32, serialPort io.ReadWriteCloser) error {
// 	if serialPort == nil {
// 		return nil
// 	}

// 	// Package argument for transmission
// 	bufOut := new(bytes.Buffer)
// 	err := binary.Write(bufOut, binary.LittleEndian, argument)
// 	if err != nil {
// 		return err
// 	}

// 	// Transmit command and argument down the pipe.
// 	for _, v := range [][]byte{[]byte{command}, bufOut.Bytes()} {
// 		_, err = serialPort.Write(v)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
