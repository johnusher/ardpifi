package oled

import (
	"image"

	log "github.com/sirupsen/logrus"
)

type mockOLED struct {
	// b          []byte
	// sync.Mutex // protects byte buffer
	log *log.Entry
}

func (m *mockOLED) On() error {
	m.log.Info("On")
	return nil
}

func (m *mockOLED) Off() error {
	m.log.Info("Off")
	return nil
}

func (m *mockOLED) Clear() error {
	m.log.Info("Clear")
	return nil
}

func (m *mockOLED) SetPixel(x, y int, v byte) error {
	m.log.Infof("SetPixel: %d %d %v", x, y, v)
	return nil
}

func (m *mockOLED) SetImage(x, y int, img image.Image) error {
	m.log.Infof("SetImage: %d %d", x, y)
	return nil
}

func (m *mockOLED) Draw() error {
	m.log.Info("Draw")
	return nil
}

func (m *mockOLED) EnableScroll(startY, endY int) error {
	m.log.Infof("EnableScroll: %d %d", startY, endY)
	return nil
}

func (m *mockOLED) DisableScroll() error {
	m.log.Info("DisableScroll")
	return nil
}

func (m *mockOLED) Width() int {
	return 128
}

func (m *mockOLED) Height() int {
	return 64
}

func (m *mockOLED) Close() error {
	m.log.Info("Close")
	return nil
}

func (m *mockOLED) ShowText(img *image.RGBA, line int, txtLabel string) {
	m.log.Infof("ShowText: %d %s", line, txtLabel)
}

func (m *mockOLED) AddGesture(img *image.RGBA, letterImage [28][28]byte) {
	m.log.Infof("AddGesture")
}

// AddGesture(img *image.RGBA, letterImage [28][28]byte)
