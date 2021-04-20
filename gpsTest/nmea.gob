package main

import (
    "errors"
    "fmt"
    "math"
    "strconv"
    "strings"
)

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