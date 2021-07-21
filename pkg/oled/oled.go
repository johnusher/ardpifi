package oled

// https://pkg.go.dev/github.com/goiot/devices@v0.0.0-20160708214026-09d1226fc8ea/monochromeoled
// 128 x 64 pixel oled screen
// Package monochromeoled contains an Adafruit Monochrome OLED (SSD1306) display driver.

import (
	"image"
	"image/color"

	"github.com/goiot/devices/monochromeoled"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/io/i2c/driver"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// OLED defines the interface for either a oled or mockOLED
type OLED interface {
	On() error
	// Off turns off the display if it is on.
	Off() error
	// Clear clears the entire display.
	Clear() error
	SetPixel(x, y int, v byte) error
	// SetImage draws an image on the display buffer starting from x, y.
	// A call to Draw is required to display it on the OLED display.
	SetImage(x, y int, img image.Image) error
	// Draw draws the intermediate pixel buffer on the display.
	// See SetPixel and SetImage to mutate the buffer.
	Draw() error
	// StartScroll starts scrolling in the horizontal direction starting from
	// startY column to endY column.
	EnableScroll(startY, endY int) error
	// StopStrolls stops the scrolling on the display.
	DisableScroll() error
	// Width returns the display width.
	Width() int
	// Height returns the display height.
	Height() int
	// Close closes the display.
	Close() error

	// ShowText displays text on specific line to OLED. This is an addition to the
	// existing monochromeoled.OLED functionality.
	ShowText(img *image.RGBA, line int, txtLabel string)
	AddGesture(img *image.RGBA, letterImage [28][28]byte)
	// func (o *oled) AddGesture(img *image.RGBA, letterImage [28][28]byte) {

}

// oled wraps a monochromeoled.OLED, and provides ShowText technology.
type oled struct {
	*monochromeoled.OLED
}

func Open(o driver.Opener, mock bool) (OLED, error) {
	if mock {
		return &mockOLED{
			log: log.WithFields(log.Fields{
				"mock": "oled",
			}),
		}, nil
	}

	mOLED, err := monochromeoled.Open(o)
	if err != nil {
		return nil, err
	}

	return &oled{OLED: mOLED}, nil
}

// ShowText displays text on specific line to OLED
func (o *oled) ShowText(img *image.RGBA, line int, txtLabel string) {
	clearLine(img, line)
	addLabel(img, 0, line, txtLabel)
	o.SetImage(0, 0, img)
	o.Draw()
}

// AddGesture adds a 28x28 image to lower part of screen
func (o *oled) AddGesture(img *image.RGBA, letterImage [28][28]byte) {
	w := 28 //
	M := 30 //offset

	// lineOffset := (line - 1) * 10
	col1 := color.RGBA{200, 100, 0, 255}
	col0 := color.RGBA{0, 0, 0, 255}

	for y := M; y < M+w; y++ {
		for x := 1; x < w; x++ {
			// if letterImage[y-M][x] == 1 {
			if letterImage[x][y-M] == 1 {
				img.Set(x, y, col1)

			} else {
				img.Set(x, y, col0)

			}
		}
	}

	o.SetImage(0, 0, img)
	o.Draw()
}

// addLabel adds a text label to image
func addLabel(img *image.RGBA, x, line int, label string) {
	lineOffset := (line) * 10

	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(lineOffset * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)

}

// clearLine clears a line on the OLDE screen
func clearLine(img *image.RGBA, line int) {
	col := color.RGBA{0, 0, 0, 0}
	// w := d.Width()
	w := 128 // pixel width of OLED screen
	lineOffset := (line - 1) * 10
	for y := 1; y < 10; y++ {
		for x := 1; x < w; x++ {
			img.Set(x, y+lineOffset, col)
		}
	}

}
