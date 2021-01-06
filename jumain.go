// public ardpifi

// to check which I2C device are connected run i2cdetect -y 1 (apt install i2c-tools )

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
	log "github.com/sirupsen/logrus"

	"github.com/johnusher/ardpifi/pkg/keyboard"
)

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	keys := make(chan rune)

	kb, err := keyboard.Init(keys)
	if err != nil {
		log.Errorf("failed to initialize keyboard: %s", err)
		return
	}
	// defer kb.Close()
	go kb.Run()

	for {
		// listen on the keys channel for key presses
		select {
		case key, more := <-keys:
			if !more {
				log.Infof("keyboard listener closed")

				// termbox closed, block until ctrl-c is called
				<-stop

				log.Infof("exiting")
				return
			}
			log.Infof("key pressed: %s / %d / 0x%X / 0%o", string(key), key, key, key)
			// n, err := i2c.WriteBytes([]byte{0x1, 0xF3})
		}
	}

	// everything below here is unreachable due to the event loop above that exits

	defer logger.FinalizeLogger()
	fmt.Println("started i2c'ing")

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
	n, err := i2c.WriteBytes([]byte{0x1, 0xF3})
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
