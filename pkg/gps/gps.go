package gps

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
	Lat        float64
	Long       float64
	FixQuality uint16 // Horizontal Dilution of Precision (HDOP). Relative accuracy of horizontal position. 1 = ideal, >20 = poor
}

type gps struct {
	gps        chan<- GPSMessage
	SerialPort io.ReadWriteCloser // i was trying to pass this out of init and then back in to run
}

func Init(gpsChan chan<- GPSMessage, noHardware bool) (GPS, error) {
	if noHardware {
		return initMockGPS(gpsChan)
	}

	return initGPS(gpsChan)
}

func initGPS(gpsChan chan<- GPSMessage) (GPS, error) { // !!! not sure about this!

	// NB I wanted to open the serial port here, then return the serial port, and use it in the run()

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
			if gps.fixQuality == "1" || gps.fixQuality == "2" {
				latitude, _ := gps.GetLatitude()
				longitude, _ := gps.GetLongitude()

				// NB latitude and longitude are strings so need to cpnvet to float:

				latitudeF, _ := strconv.ParseFloat(latitude, 64)
				longitudeF, _ := strconv.ParseFloat(longitude, 64)
				fixQuality, _ := strconv.ParseInt(gps.fixQuality, 10, 16)

				g.gps <- GPSMessage{
					Lat:        latitudeF,
					Long:       longitudeF,
					FixQuality: uint16(fixQuality),
				}

				// log.Infof("LAtitude =  %s. Longitude = %s", latitude, longitude)
				// log.Infof("fixQuality =  %s. ", fixQuality)

				// fmt.Println(latitude + "," + longitude)
				// result, _ := geocoder.reverse(Position{Latitude: latitude, Longitude: longitude})

			} else {
				// fmt.Println("no gps fix available")
				log.Infof("low fixQuality %s \n", gps.fixQuality)
			}
			time.Sleep(2 * time.Second)
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
	if tokens[0] == "$GPGGA" {
		return NMEA{
			fixTimestamp:       tokens[1],
			latitude:           tokens[2],
			latitudeDirection:  tokens[3],
			longitude:          tokens[4],
			longitudeDirection: tokens[5],
			fixQuality:         tokens[6],
			satellites:         tokens[7],
		}, nil
	}
	return NMEA{}, errors.New("unsupported nmea string")
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
