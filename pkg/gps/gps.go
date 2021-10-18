package gps

// return GS cooridinates and HDOP.
// where HDOP:
// <1	Ideal	Highest possible confidence level to be used for applications demanding the highest possible precision at all times.
// 1-2	Excellent	At this confidence level, positional measurements are considered accurate enough to meet all but the most sensitive applications.
// 2-5	Good	Represents a level that marks the minimum appropriate for making accurate decisions. Positional measurements could be used to make reliable in-route navigation suggestions to the user.
// 5-10	Moderate	Positional measurements could be used for calculations, but the fix quality could still be improved. A more open view of the sky is recommended.

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/jacobsa/go-serial/serial"
)

type GPS interface {
	Run() error
	Close() error
}

type GPSMessage struct {
	Lat  float64
	Long float64
	// FixQuality uint16 // Horizontal Dilution of Precision (HDOP). Relative accuracy of horizontal position. 1 = ideal, >20 = poor
	HDOP float64 // Horizontal Dilution of Precision (HDOP). Relative accuracy of horizontal position. 0.0 to 9.9
}

type gps struct {
	gps        chan<- GPSMessage
	SerialPort io.ReadWriteCloser
}

func Init(gpsChan chan<- GPSMessage, mock bool) (GPS, error) {
	if mock {
		return initMockGPS(gpsChan)
	}

	return initGPS(gpsChan)
}

func initGPS(gpsChan chan<- GPSMessage) (GPS, error) {

	options := serial.OpenOptions{
		PortName:        "/dev/ttyS0",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	serialPort, err := serial.Open(options)
	if err != nil {
		log.Errorf("serial.Open: %v", err)
		return nil, err
	}

	return &gps{
		gpsChan,
		serialPort,
	}, nil
}

func (g *gps) Close() error {
	return g.SerialPort.Close()
}

func (g *gps) Run() error {

	reader := bufio.NewReader(g.SerialPort)
	scanner := bufio.NewScanner(reader)

	// log.Infof("Started GPS read with port %s", g.SerialPort)

	for scanner.Scan() {

		gps, err := ParseNMEALine(scanner.Text())
		if err == nil {
			// if gps.fixQuality == "1" || gps.fixQuality == "2" {
			latitude, _ := gps.GetLatitude()
			longitude, _ := gps.GetLongitude()
			hdop := gps.GetHorizontalDilution()

			// NB latitude and longitude are strings so need to cpnvet to float:

			latitudeF, _ := strconv.ParseFloat(latitude, 64)
			longitudeF, _ := strconv.ParseFloat(longitude, 64)
			// fixQuality, _ := strconv.ParseInt(gps.fixQuality, 10, 16)
			hdopF, _ := strconv.ParseFloat(hdop, 64)

			g.gps <- GPSMessage{
				Lat:  latitudeF,
				Long: longitudeF,
				// FixQuality: uint16(fixQuality),
				HDOP: hdopF,
			}

			// log.Infof("LAtitude =  %s. Longitude = %s", latitude, longitude)
			// log.Infof("fixQuality =  %s. ", fixQuality)

			// fmt.Println(latitude + "," + longitude)
			// result, _ := geocoder.reverse(Position{Latitude: latitude, Longitude: longitude})

			// } else {
			// 	// fmt.Println("no gps fix available")
			// 	log.Infof("low fixQuality %s \n", gps.fixQuality)
			// }
			// time.Sleep(1 * time.Second)

			time.Sleep(123 * time.Millisecond)

		} else {
			// log.Infof("ParseNMEALine error, %s", err)
		}
	}

	return nil
}

type NMEA struct {
	fixTimestamp       string
	latitude           string
	latitudeDirection  string
	longitude          string
	longitudeDirection string
	fixQuality         string
	satellites         string
	horizontalDilution string
	antennaAltitude    string
	antennaHeight      string
	updateAge          string
}

func ParseNMEALine(line string) (NMEA, error) {
	tokens := strings.Split(line, ",")
	if len(tokens) < 8 {
		return NMEA{}, fmt.Errorf("unsupported nmea string, expected 8 tokens got %d: %s", len(tokens), line)
	}
	// if tokens[0] != "$GPGGA" || tokens[0] != "$GNGSA" || tokens[0] != "$GNGLL" {
	// 	return NMEA{}, fmt.Errorf("unsupported nmea string: %s", line)
	// }

	// NB untested for GPGGA
	if tokens[0] == "$GNGGA" || tokens[0] == "$GPGGA" {
		return NMEA{
			fixTimestamp:       tokens[1],
			latitude:           tokens[2],
			latitudeDirection:  tokens[3],
			longitude:          tokens[4],
			longitudeDirection: tokens[5],
			fixQuality:         tokens[6],
			satellites:         tokens[7],
			horizontalDilution: tokens[8], // Horizontal Dilution of Precision (HDOP) 1.0 to 9.9
		}, nil
	} else {
		return NMEA{}, fmt.Errorf("unsupported nmea string: %s", line)

	}

	// if tokens[0] != "$GNGSA" || tokens[0] != "$GNGLL" {
	// 	return NMEA{}, fmt.Errorf("unsupported nmea string: %s", line)
	// }

	// if tokens[0] == "$GNGLL" {
	// 	return NMEA{
	// 		// fixTimestamp:       tokens[1],
	// 		latitude:           tokens[1],
	// 		latitudeDirection:  tokens[2],
	// 		longitude:          tokens[3],
	// 		longitudeDirection: tokens[4],
	// 		// fixQuality:         "1",
	// 		// fixQuality:         tokens[6],
	// 		// satellites:         tokens[7],
	// 	}, nil
	// } else if tokens[0] == "$GNGSA" {
	// 	// https://www.hemispheregnss.com/technical-resource-manual/Import_Folder/GNGSA_Message.htm
	// 	return NMEA{
	// 		// fixTimestamp:       tokens[1],
	// 		// latitude:           tokens[1],
	// 		// latitudeDirection:  tokens[2],
	// 		// longitude:          tokens[3],
	// 		// longitudeDirection: tokens[4],
	// 		horizontalDilution: tokens[16], // Horizontal Dilution of Precision (HDOP) 1.0 to 9.9
	// 		// fixQuality:         tokens[6],
	// 		// satellites:         tokens[7],
	// 	}, nil
	// 	// } else if tokens[0] == "$GPGGA" {
	// } else {
	// 	// https://www.hemispheregnss.com/technical-resource-manual/Import_Folder/GNGSA_Message.htm
	// 	return NMEA{
	// 		// fixTimestamp:       tokens[1],
	// 		// latitude:           tokens[2],
	// 		// latitudeDirection:  tokens[3],
	// 		// longitude:          tokens[4],
	// 		// longitudeDirection: tokens[5],
	// 		// fixQuality:         tokens[6],
	// 		// satellites:         tokens[7],
	// 	}, nil
	// }

}

func ParseDegrees(value string, direction string) (string, error) {
	if value == "" || direction == "" {
		return "", errors.New("the location and / or direction value does not exist")
	}
	lat, _ := strconv.ParseFloat(value, 64)
	degrees := math.Floor(lat / 100)
	minutes := ((lat / 100) - math.Floor(lat/100)) * 100 / 60
	decimal := degrees + minutes
	if direction == "W" || direction == "S" {
		decimal *= -1
	}
	return fmt.Sprintf("%.6f", decimal), nil
}

func (nmea NMEA) GetLatitude() (string, error) {
	return ParseDegrees(nmea.latitude, nmea.latitudeDirection)
}

func (nmea NMEA) GetLongitude() (string, error) {
	return ParseDegrees(nmea.longitude, nmea.longitudeDirection)
}

func (nmea NMEA) GetHorizontalDilution() string {
	// Horizontal Dilution of Precision (HDOP) 1.0 to 9.9
	return nmea.horizontalDilution
}
