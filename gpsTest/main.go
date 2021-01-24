package main

import (
    "bufio"
    "flag"
    "fmt"
    "log"
    "time"

    "github.com/jacobsa/go-serial/serial"
)

func main() {
    options := serial.OpenOptions{
        PortName:        "/dev/ttyS0",
        BaudRate:        9600,
        DataBits:        8,
        StopBits:        1,
        MinimumReadSize: 4,
    }
    serialPort, err := serial.Open(options)
    if err != nil {
        log.Fatalf("serial.Open: %v", err)
    }
    defer serialPort.Close()
    reader := bufio.NewReader(serialPort)
    scanner := bufio.NewScanner(reader)
    geocoder := Geocoder{AppId: "APP-ID-HERE", AppCode: "APP-CODE-HERE"}
    for scanner.Scan() {
        gps, err := ParseNMEALine(scanner.Text())
        if err == nil {
            if gps.fixQuality == "1" || gps.fixQuality == "2" {
                latitude, _ := gps.GetLatitude()
                longitude, _ := gps.GetLongitude()
                fmt.Println(latitude + "," + longitude)
                result, _ := geocoder.reverse(Position{Latitude: latitude, Longitude: longitude})
                if len(result.Response.View) > 0 && len(result.Response.View[0].Result) > 0 {
                    fmt.Println(result.Response.View[0].Result[0].Location.Address.Label)
                } else {
                    fmt.Println("no address estimates found for the position")
                }
            } else {
                fmt.Println("no gps fix available")
            }
            time.Sleep(2 * time.Second)
        }
    }
}
