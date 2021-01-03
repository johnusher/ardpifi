package main

import (
	"fmt"
	"i2c"
	"log"

	"github.com/d2r2/go-logger"
)

func main() {

	defer logger.FinalizeLogger()
	fmt.Println("started i2c'ing")

	// Create new connection to I2C bus on 2 line with address 0x27
	i2c, err := i2c.NewI2C(0x08, 1)
	if err != nil {
		log.Fatal(err)
	}
	// Free I2C connection on exit
	defer i2c.Close()

	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	// i2c.Write([]byte{0x1, 0xF3})
	// Here goes code specific for sending and reading data
	// to and from device connected via I2C bus, like:
	// x, err := i2c.Write([]byte{0x1, 0xF3})
	// if err != nil {
	// 	log.Fatal(err)
	// }
	//i2c.Write([]byte{0x1, 0xF3})
	i2c.WriteBytes([]byte{0x1, 0xF3})

	// Uncomment/comment next line to suppress/increase verbosity of output
	logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

	fmt.Println("ended")

}
