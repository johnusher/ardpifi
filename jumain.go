// public ardpifi

// to check which I2C device are connected run i2cdetect -y 1 (apt install i2c-tools )

package main

import (
	"fmt"
	"log"

	"github.com/d2r2/go-i2c"
	"github.com/d2r2/go-logger"
)

func main() {

	defer logger.FinalizeLogger()
	fmt.Println("started i2c'ing")

	// Create new connection to Arduino:
	// I2C bus on line 1 with address 0x08
	i2c, err := i2c.NewI2C(0x08, 1)
	if err != nil {
		log.Fatal(err)
	}
	// Free I2C connection on exit
	defer i2c.Close()

	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	// // this next bit does not work:!!
	// from https://github.com/d2r2/go-i2c
	// ....
	// // Here goes code specific for sending and reading data
	// // to and from device connected via I2C bus, like:
	// _, err := i2c.Write([]byte{0x1, 0xF3})
	// if err != nil { log.Fatal(err) }
	// ....

	// write data to I2C
	i2c.WriteBytes([]byte{0x1, 0xF3})

	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	fmt.Println("ended")

}
