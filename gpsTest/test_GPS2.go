// JU_led_mesh

// based on https://github.com/siggy/ledmesh/blob/master/main.go

package main

import (
	"flag"
	"fmt"
	"os/signal"
	"syscall"

	"os"
	"time"

	// "github.com/qinxin0720/lcd1602"

	// i2c "github.com/d2r2/go-i2c"

	"github.com/johnusher/ardpifi/pkg/gps"

	// "github.com/johnusher/ardpifi/pkg/lcd"
	"github.com/johnusher/ardpifi/pkg/oled"

	log "github.com/sirupsen/logrus"

	"image"

	_ "image/png"

	// "github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"
)

const (
	bearingThreshold = 10 // value in degrees. if we are within eg 10 degrees pointing at another, then consider it a lock on
	batPort          = 4200
	// msgSize   = net.IPv4len + 4 // IP + uint32
	interval  = 1 * time.Second
	ifaceName = "bat0" // rpi
	// ifaceName = "en0" // pc

	batBcast   = "172.27.255.255"
	localBcast = "127.0.0.1"

	// piTTL defines how long we wait to expire a PI if we haven't received a
	// message from it.
	piTTL = 30 * time.Second

	Pi = 3.14159265358979323846264338327950288419716939937510582097494459 // pi https://oeis.org/A000796

	magicByte = ("BA")

	messageTypeGPS   = 0
	messageTypeDuino = 1

	raspiIDEveryone = "00"
)

// ChatRequest is ChatRequest, stop telling me about comments
type ChatRequest struct {
	Latf     float64
	Longf    float64
	HDOPf    float64
	ID       string
	Key      rune
	PointDir int64
}

type chatRequestWithTimestamp struct {
	ChatRequest
	lastMessageReceived time.Time
}

// String satisfies the Stringer interface
func (c ChatRequest) String() string {
	return fmt.Sprintf("id: %s, coords: (%f, %f), HDOP: %.2f", c.ID, c.Latf, c.Longf, c.HDOPf)
}

// String satisfies the Stringer interface
func (c chatRequestWithTimestamp) String() string {
	return fmt.Sprintf("%s, age: %s]", c.ChatRequest, time.Now().Sub(c.lastMessageReceived))
}

func main() {
	noGPS := flag.Bool("no-gps", false, "run without gps")
	noOLED := flag.Bool("no-oled", false, "run without oled display")

	logLevel := flag.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")

	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("failed to parse log level [%s]: %s", *logLevel, err)
		return
	}
	log.SetLevel(level)

	// OLED:

	oled, err := oled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"}, *noOLED)
	if err != nil {
		panic(err)
	}
	defer oled.Close()

	// load png and display on OLED
	rc, err := os.Open("./maxi.png")

	if err != nil {
		panic(err)
	}
	defer rc.Close()

	m, _, err := image.Decode(rc)
	if err != nil {
		panic(err)
	}

	// clear the display before putting on anything
	if err := oled.Clear(); err != nil {
		panic(err)
	}

	if err := oled.SetImage(0, 0, m); err != nil {
		panic(err)
	}
	if err := oled.Draw(); err != nil {
		panic(err)
	}

	// init GPS module:
	gpsChan := make(chan gps.GPSMessage)
	g, err := gps.Init(gpsChan, *noGPS)
	if err != nil {
		log.Errorf("failed to initialize gps: %s", err)
		return
	}
	defer g.Close()

	// go forth

	go g.Run()

	errs := make(chan error)

	// clear the OLED
	if err := oled.Clear(); err != nil {
		panic(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 128, 64))

	go func() {
		errs <- broadcastLoop(gpsChan, img, oled)
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// // block until ctrl-c or one of the loops returns an error
	select {
	case <-stop:
	case <-errs:
	}
}

func broadcastLoop(gpsCh <-chan gps.GPSMessage, img *image.RGBA, oled oled.OLED) error {
	log.Info("Starting broadcast loop")

	// this is for local messages, eg key-presses, GPS update, pointing direction

	gpsMessage := gps.GPSMessage{}

	more := false

	for {
		select {

		case gpsMessage, more = <-gpsCh:

			// received GPS from local GPS module

			if !more {
				log.Infof("gps channel closed\n")
				log.Infof("exiting")
				return nil
			}

			log.Infof("gpsMessage: %v \n", gpsMessage)

			// 8 bytes: Lat
			// 8 bytes: Long
			// 8 bytes: HDOP

			if gpsMessage.Lat != 0.0 {

				// GPSmsgSize := 24                       // 24 bytes for  a gps message

				// GPSmsgSize :=    // 32 bytes for  a gps message

				// copy(messageOut[0:2], magicByte)
				// messageOut[2] = uint8(GPSmsgSize)
				// copy(messageOut[3:5], raspID)

				// whoFor := raspiIDEveryone // message for everyone
				// copy(messageOut[5:7], whoFor)

				// messageType := messageTypeGPS // GPS
				// messageOut[7] = uint8(messageType)

				// log.Infof("key pressed: %s / %d / 0x%X / 0%o \n", string(key), key, key, key)

				// // now split the float64 lat and long values into bytes and shove them in the message
				// binary.LittleEndian.PutUint64(messageOut[8:16], math.Float64bits(gpsMessage.Lat))
				// binary.LittleEndian.PutUint64(messageOut[16:24], math.Float64bits(gpsMessage.Long))
				// binary.LittleEndian.PutUint64(messageOut[24:32], math.Float64bits(gpsMessage.HDOP))

				// todo: send HDOP!!

			}

			// OLED display:
			if gpsMessage.HDOP != 0.0 {
				msgP := fmt.Sprintf("HDOP = %.2f", gpsMessage.HDOP)
				oled.ShowText(img, 6, msgP)
			}

		}
	}
}
