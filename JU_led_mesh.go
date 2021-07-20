// JU_led_mesh

// based on https://github.com/siggy/ledmesh/blob/master/main.go

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	// "github.com/qinxin0720/lcd1602"

	// i2c "github.com/d2r2/go-i2c"
	"github.com/johnusher/ardpifi/pkg/acc"
	"github.com/johnusher/ardpifi/pkg/gps"
	"github.com/johnusher/ardpifi/pkg/iface"
	"github.com/johnusher/ardpifi/pkg/keyboard"

	// "github.com/johnusher/ardpifi/pkg/lcd"
	"github.com/johnusher/ardpifi/pkg/oled"
	"github.com/johnusher/ardpifi/pkg/port"
	"github.com/johnusher/ardpifi/pkg/readBATMAN"
	"github.com/johnusher/ardpifi/pkg/web"

	log "github.com/sirupsen/logrus"
	"github.com/tarm/serial"

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
	raspID := flag.String("rasp-id", "r1", "unique raspberry pi ID") // we need to make this 2 bytes!
	webAddr := flag.String("web-addr", ":8080", "address to serve web on")
	noBatman := flag.Bool("no-batman", false, "run without batman network")
	noDuino := flag.Bool("no-duino", false, "run without arduino")
	noGPS := flag.Bool("no-gps", false, "run without gps")
	noOLED := flag.Bool("no-oled", false, "run without oled display")
	noACC := flag.Bool("no-acc", false, "run without Bosch accelerometer")

	logLevel := flag.String("log-level", "info", "log level, must be one of: panic, fatal, error, warn, info, debug, trace")

	flag.Parse()

	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		log.Errorf("failed to parse log level [%s]: %s", *logLevel, err)
		return
	}
	log.SetLevel(level)

	// make raspID into 2 bytes: take first 2 letter if needed:

	*raspID = (*raspID)[0:2]

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

	web := web.InitWeb(*webAddr)
	log.Infof("web: %+v", web)

	bcastIP := net.ParseIP(batBcast)
	if *noBatman {
		bcastIP = net.ParseIP(localBcast)
	}

	// Find the device that represents the arduino serial
	// connection. NB this is kinda janky- we should have a system to robustly detect a duino,
	// eg if we dont find one, then re-insert the duino USb cable and note which ports are new

	// c := &serial.Config{Name: findArduino(), Baud: 9600, ReadTimeout: time.Second * 1}
	// c := &serial.Config{Name: findArduino(), Baud: 19200, ReadTimeout: time.Second * 1}
	c := &serial.Config{Name: findArduino(), Baud: 115200, ReadTimeout: time.Second * 1}

	duino, err := port.OpenPort(c, *noDuino)
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

	i, err := iface.InterfaceByName(ifaceName, *noBatman, bcastIP)
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

	// init BATMAN:
	messages := make(chan []byte)
	bm, err := readBATMAN.Init(messages, *noBatman, bcastIP)
	if err != nil {
		log.Errorf("failed to initialize readBATMAN: %s", err)
		return
	}

	// init accelerometer module (Bosch)
	accChan := make(chan acc.ACCMessage)
	a, err := acc.Init(accChan, *noACC)
	if err != nil {
		log.Errorf("failed to initialize acc: %s", err)
		return
	}
	// defer a.Close()

	// init GPS module:
	gpsChan := make(chan gps.GPSMessage)
	g, err := gps.Init(gpsChan, *noGPS)
	if err != nil {
		log.Errorf("failed to initialize gps: %s", err)
		return
	}
	defer g.Close()

	// go forth
	go kb.Run()
	go bm.Run()
	go g.Run()
	go a.Run()

	errs := make(chan error)

	// clear the OLED
	if err := oled.Clear(); err != nil {
		panic(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 128, 64))

	go func() {
		errs <- messageLoop(messages, accChan, duino, *raspID, img, oled, web, bcastIP, bm)
	}()
	go func() {
		errs <- broadcastLoop(keys, gpsChan, duino, *raspID, bcastIP, bm, img, oled)
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

// messageLoop receives incoming messages
// 2 bytes: <2 magic bytes>
// 1 byte:  <total message length, bytes>
// 2 bytes: <sender ID = 2 bytes, (IP?)>
// 2 bytes: <who For = 2 bytes, (0= everyone, or ID of)>
// 1 byte:  <message type (0=gps, 1=duino command, 2=gesture type)>
// N bytes: <message, >0 bytes>
func messageLoop(messages <-chan []byte, accCh <-chan acc.ACCMessage, duino port.Port, raspID string, img *image.RGBA, oled oled.OLED, web *web.Web, bcastIP net.IP, bm *readBATMAN.ReadBATMAN) error {
	log.Info("Starting message loop")
	// listen on the keys channel for key presses AND listen for new BATMAN message
	// allPIs keeps track of the last message received from each PI, keyed by
	// raspID
	allPIs := map[string]chatRequestWithTimestamp{}
	accMessage := acc.ACCMessage{}
	bcast := &net.UDPAddr{Port: batPort, IP: bcastIP}

	more := false

	for {

		select {

		case accMessage, more = <-accCh:

			// received message from BNo055 module.
			// eg bearing, ie NSEW direction we are pointing
			if !more {
				log.Infof("acc channel closed\n")
				log.Infof("exiting")
				return nil
			}

			// accMessage.Bearing is "pointing direction" of self
			// note we also use the term "Bearing" for the relative direction between pis
			// from the mock, accMessage.Bearing was always rand * 0.0001, so never
			// larger than 0.0001, so always rounded to zero
			bearingI := int64(math.Round(accMessage.Bearing))

			// save to self
			crwt, _ := allPIs[raspID]
			crwt.PointDir = bearingI // trying to save here but later, it just pulls a zero!
			allPIs[raspID] = crwt

			// msgP := fmt.Sprintf("Pointing direction = %d", bearingI)
			// log.Infof(msgP)

			// OLED display:
			msgP := fmt.Sprintf("Pointing = %d", bearingI)
			oled.ShowText(img, 1, msgP)

		case message, _ := <-messages:

			magicBytesRx := string(message[0:2]) // combine the magicBytes

			if magicBytesRx != magicByte {
				log.Errorf("Received magicBytes %s, expected %s", string(magicBytesRx), magicByte)
				continue
			}

			// messageLength := uint8(message[2])  // todo: check message is correct length!

			senderID := string(message[3:5]) // this is length of raspID = 2 bytes

			whoFor := string(message[5:7]) // whoFor is also length of raspID = 2 bytes

			messageType := message[7]

			if senderID == raspID {
				// // senderID and raspID should both be two bytes, ie two characters
				// if messageType == messageTypeDuino {
				// 	// write to duino:
				// 	// this is currently kinda redundant, ie whether the message is from self or other, we send it to duino
				// 	// ... but one day we may send a different message for self-sent message

				// 	// first unpack the message:
				// 	duinoMessage := message[8] // we should maybe look at total message legnth and combine other bytes if longer than 7

				// 	// write to duino:
				// 	duino.Flush()
				// 	_, err := duino.Write([]byte(string(duinoMessage)))

				// 	if err != nil {
				// 		log.Errorf("3. failed to write to serial port: %s", err)
				// 		//return err
				// 	}
				// 	duino.Flush()
				// }

			} else {

				if whoFor == raspiIDEveryone || whoFor == raspID { // the strcmp with whoFor doesnt work!!
					// if message[6] == 0 || whoFor == raspID { // message[6] == 0  means for everyone.
					// message is not sent by self and is for everyone or for me

					if messageType == messageTypeDuino {

						// duino command: send straight to duino
						// first unpack the message:
						duinoMessage := message[8] // we should maybe look at total message legnth and combine other bytes if longer than 7

						// write to duino:
						duino.Flush()
						_, err := duino.Write([]byte(string(duinoMessage)))

						if err != nil {
							log.Errorf("3. failed to write to serial port: %s", err)
							//return err
						}
						duino.Flush()

						log.Infof("key from other %s \n", (string(duinoMessage)))

						// OLED display:
						OLEDmsg := fmt.Sprintf("received:  %s", (string(duinoMessage)))
						oled.ShowText(img, 2, OLEDmsg)
					}
				}
			}

			// now we update our list of active pis on the network:
			now := time.Now()

			crwt, ok := allPIs[senderID]
			if !ok {
				log.Infof("new PI detected: %+v", senderID)
			}
			crwt.lastMessageReceived = now
			crwt.ID = senderID

			// now do some general house-keeping, set device IDs on the network etc:

			if messageType == messageTypeGPS {
				// gps package

				// Received Lattitude is a float 64 in message bytes 8:15
				// Received Long is a float 64 in message bytes 16:23

				rxLatBytes := message[8:16]
				bits := binary.LittleEndian.Uint64(rxLatBytes)
				rxLatFloat := math.Float64frombits(bits)

				crwt.Latf = rxLatFloat

				rxLongBytes := message[16:24]
				bits = binary.LittleEndian.Uint64(rxLongBytes)
				rxLongFloat := math.Float64frombits(bits)

				crwt.Longf = rxLongFloat

				rxHDOPBytes := message[24:32]
				bits = binary.LittleEndian.Uint64(rxHDOPBytes)
				rxHDOPFloat := math.Float64frombits(bits)

				crwt.HDOPf = rxHDOPFloat

				// todo: HDOP
			}

			allPIs[senderID] = crwt

			// remove any PIs we haven't heard from in a while
			for k, v := range allPIs {
				if v.lastMessageReceived.Add(piTTL).Before(now) {
					log.Infof("deleting expired pi: %+v", v)
					delete(allPIs, k)
				}
			}

			log.Infof("current PIs: %d", len(allPIs))
			for _, v := range allPIs {
				log.Infof("  %s", v)
			}

			if messageType == messageTypeGPS {
				if self, ok := allPIs[raspID]; ok && len(allPIs) > 1 {
					// we have >1 Pis, including ourself, find bearing and distance from local to each pi

					// NB should we also do this when we have a new estimate for our local GPS location?

					lat1 := self.Latf
					long1 := self.Longf

					currentPD := self.PointDir // current pointing direction of self/ this returns zeros!!

					// msgP := fmt.Sprintf("Pointing direction = %d", currentPD)
					// log.Infof(msgP)

					for piID, crwt := range allPIs {
						if piID == raspID {
							// this is ourself, skip
							continue
						}

						lat2 := crwt.Latf
						long2 := crwt.Longf

						bearing, _ := calcGPSBearing(lat1, long1, lat2, long2)
						disance := calcGPSdistance(lat1, long1, lat2, long2)
						bearingI := int64(math.Round(bearing))
						distI := int64(math.Round(disance))

						// now see if bearing to this other pi matches pointing direction of the current pi:
						bearingMistmatch := int64(1)
						if currentPD > 360-bearingThreshold && bearingI < bearingThreshold {
							bearingMistmatch = Abs(currentPD - (bearingI + 360))
						} else if bearingI > 360-bearingThreshold && currentPD < bearingThreshold {
							bearingMistmatch = Abs(bearingI - (currentPD + 360))
						} else {
							bearingMistmatch = Abs(currentPD - bearingI)
						}

						// msgP := fmt.Sprintf("currentPD, %d", currentPD)
						// log.Infof(msgP)

						// msgP = fmt.Sprintf("bearingMistmatch, %d", bearingMistmatch)
						// log.Infof(msgP)

						if bearingMistmatch < bearingThreshold {
							// we are pointing at another!!
							// send key=1 to network, to the piID. ie using broadcastLoop

							// duino message (9 bytes)
							// 2 bytes: <2 magic bytes>
							// 1 byte:  <total message length, bytes>
							// 2 bytes: <sender ID = 2 bytes, (IP?)>
							// 2 bytes: <who For = 2 bytes, (0= everyone, or ID of)>
							// 1 byte:  <message type (0=gps, 1=duino command, 2=gesture type)>
							// 1 byte:  key

							duinoMsgSize := 9                        // 23 bytes for a duino message
							messageOut := make([]byte, duinoMsgSize) // sent to batman

							copy(messageOut[0:2], magicByte)
							messageOut[2] = uint8(duinoMsgSize)
							copy(messageOut[3:5], raspID)

							// send just to the one we are pointing at:
							whoFor := crwt.ID // or should this be piID??
							copy(messageOut[5:7], whoFor)

							messageType := messageTypeDuino // duino message
							messageOut[7] = uint8(messageType)
							sendMessage := rune('1')
							messageOut[8] = uint8(sendMessage)

							_, err := bm.Conn.WriteToUDP(messageOut, bcast)

							if err != nil {
								log.Error(err)
								return err
							}

							msgP := fmt.Sprintf("We are pointing at %s", crwt.ID)
							log.Infof(msgP)
							// msgP = fmt.Sprintf("Pointing at", crwt.ID)
							// oled.ShowText(img, 5, msgP)

						}

						msg1 := fmt.Sprintf("bearing to %s = %d\xB0", crwt.ID, bearingI)
						log.Infof(msg1)
						msg2 := fmt.Sprintf("dist to %s = %d m", crwt.ID, distI)
						log.Infof(msg2)

						// msg1 = fmt.Sprintf("bearing = %d\xB0", bearingI)
						msg1 = fmt.Sprintf("bearing = %d", bearingI)
						msg2 = fmt.Sprintf("dist = %d m", distI)
						oled.ShowText(img, 3, msg1)
						oled.ShowText(img, 4, msg2)
					}
				}
			}
		}

	}
}

func broadcastLoop(keys <-chan rune, gpsCh <-chan gps.GPSMessage, duino port.Port, raspID string, bcastIP net.IP, bm *readBATMAN.ReadBATMAN, img *image.RGBA, oled oled.OLED) error {
	log.Info("Starting broadcast loop")

	// this is for local messages, eg key-presses, GPS update, pointing direction

	bcast := &net.UDPAddr{Port: batPort, IP: bcastIP}
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

			// GPS message (32 bytes)
			// 2 bytes: <2 magic bytes>
			// 1 byte:  <total message length, bytes>
			// 2 bytes: <sender ID = 2 bytes, (IP?)>
			// 2 bytes: <who For = 2 bytes, (0= everyone, or ID of)>
			// 1 byte:  <message type (0=gps, 1=duino command, 2=gesture type)>
			// 8 bytes: Lat
			// 8 bytes: Long
			// 8 bytes: HDOP

			if gpsMessage.Lat != 0.0 {

				// GPSmsgSize := 24                       // 24 bytes for  a gps message

				GPSmsgSize := 32                       // 32 bytes for  a gps message
				messageOut := make([]byte, GPSmsgSize) // sent to batman

				copy(messageOut[0:2], magicByte)
				messageOut[2] = uint8(GPSmsgSize)
				copy(messageOut[3:5], raspID)

				whoFor := raspiIDEveryone // message for everyone
				copy(messageOut[5:7], whoFor)

				messageType := messageTypeGPS // GPS
				messageOut[7] = uint8(messageType)

				// now split the float64 lat and long values into bytes and shove them in the message
				binary.LittleEndian.PutUint64(messageOut[8:16], math.Float64bits(gpsMessage.Lat))
				binary.LittleEndian.PutUint64(messageOut[16:24], math.Float64bits(gpsMessage.Long))
				binary.LittleEndian.PutUint64(messageOut[24:32], math.Float64bits(gpsMessage.HDOP))

				// todo: send HDOP!!

				_, err := bm.Conn.WriteToUDP(messageOut, bcast)
				if err != nil {
					log.Error(err)
					return err
				}

			}

			// OLED display:
			if gpsMessage.HDOP != 0.0 {
				msgP := fmt.Sprintf("HDOP = %.2f", gpsMessage.HDOP)
				oled.ShowText(img, 6, msgP)
			}

		case key, more := <-keys:
			// received local key press
			// todo: replace/ augment this with a GPIO button press

			if !more {
				oled.ShowText(img, 2, fmt.Sprintf("exiting"))
				log.Infof("keyboard listener closed\n")
				// termbox closed, block until ctrl-c is called
				log.Infof("exiting")
				return nil
			}
			log.Infof("key pressed: %s / %d / 0x%X / 0%o \n", string(key), key, key, key)

			// duino message (9 bytes)
			// 2 bytes: <2 magic bytes>
			// 1 byte:  <total message length, bytes>
			// 2 bytes: <sender ID = 2 bytes, (IP?)>
			// 2 bytes: <who For = 2 bytes, (0= everyone, or ID of)>
			// 1 byte:  <message type (0=gps, 1=duino command, 2=gesture type)>
			// 1 byte:  key

			duinoMsgSize := 9                        // 23 bytes for a duino message
			messageOut := make([]byte, duinoMsgSize) // sent to batman

			copy(messageOut[0:2], magicByte)
			messageOut[2] = uint8(duinoMsgSize)
			copy(messageOut[3:5], raspID)

			whoFor := raspiIDEveryone // everyone
			copy(messageOut[5:7], whoFor)

			messageType := messageTypeDuino // duino message
			messageOut[7] = uint8(messageType)

			messageOut[8] = uint8(key)

			_, err := bm.Conn.WriteToUDP(messageOut, bcast)

			if err != nil {
				log.Error(err)
				return err
			}

			// NB now we send message ot duino after we have received it on the net- this way we sync with other duinos better
			// // write to duino: NB maybe insert a wait before here so all pi's send the new duino command at a similar time
			// _, err = duino.Write([]byte(string(key)))
			// if err != nil {
			// 	log.Errorf("2. failed to write to serial port: %s", err)
			// 	return err
			// }

			// OLED display:
			oled.ShowText(img, 2, fmt.Sprintf("key pressed: %s", string(key)))

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

func calcGPSBearing(lat1 float64, long1 float64, lat2 float64, long2 float64) (float64, error) {

	// find bearing between two decimal GPS coordinates

	// if lat1 == "" || long1 == "" {
	// 	return "", errors.New("lat1 or lat2 value does not exist")
	// }
	// lat, _ := strconv.ParseFloat(value, 64)
	// degrees := math.Floor(lat / 100)
	// minutes := ((lat / 100) - math.Floor(lat/100)) * 100 / 60
	// decimal := degrees + minutes

	// if we are stradling the equartor or the Prime Meridian, we may have a problem!!
	// todo: impliement
	// if direction == "W" || direction == "S" {
	//     decimal *= -1
	// }

	dy := lat2 - lat1
	dx := math.Cos(Pi/180.0*lat1) * (long2 - long1)
	angle := math.Atan2(dy, dx)
	angle = angle / Pi * 180.0

	if angle < 0 {
		angle = 360 + angle

	}
	// return int(math.Round(angle)), nil
	return angle, nil
}

// distance between two points.
// from https://gist.github.com/cdipaolo/d3f8db3848278b49db68

// haversin(θ) function
func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

// Distance function returns the distance (in meters) between two points of
//     a given longitude and latitude relatively accurately (using a spherical
//     approximation of the Earth) through the Haversin Distance Formula for
//     great arc distance on a sphere with accuracy for small distances
//
// point coordinates are supplied in degrees and converted into rad. in the func
//
// distance returned is METERS
// http://en.wikipedia.org/wiki/Haversine_formula
func calcGPSdistance(lat1, lon1, lat2, lon2 float64) float64 {
	// convert to radians
	// must cast radius as float to multiply later
	var la1, lo1, la2, lo2, r float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	r = 6378100 // Earth radius in METERS

	// calculate
	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)

	return 2 * r * math.Asin(math.Sqrt(h))
}

// Abs returns the absolute value of x.
func Abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}
