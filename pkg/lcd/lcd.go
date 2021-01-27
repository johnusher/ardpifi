package lcd

import (
	"sync"

	device "github.com/d2r2/go-hd44780"
	"github.com/d2r2/go-i2c"
	log "github.com/sirupsen/logrus"
)

type LCD interface {
	ShowMessage(text string, options device.ShowOptions) error
	TestWriteCGRam() error
	BacklightOn() error
	BacklightOff() error
	Clear() error
	Home() error
	SetPosition(line, pos int) error
	Command(cmd byte) error
}

type mockLCD struct {
	b          []byte
	sync.Mutex // protects byte buffer
	log        *log.Entry
}

func New(i2c *i2c.I2C, mock bool) (LCD, error) {
	if mock {
		return &mockLCD{
			log: log.WithFields(log.Fields{
				"mock": "lcd",
			}),
		}, nil
	}

	return device.NewLcd(i2c, device.LCD_16x2)
}

func (m *mockLCD) ShowMessage(text string, options device.ShowOptions) error {
	m.log.Infof("ShowMessage(%s, %+v)", text, options)
	return nil
}

func (m *mockLCD) TestWriteCGRam() error {
	m.log.Info("TestWriteCGRam")
	return nil
}

func (m *mockLCD) BacklightOn() error {
	m.log.Info("BacklightOn")
	return nil
}

func (m *mockLCD) BacklightOff() error {
	m.log.Info("BacklightOff")
	return nil
}

func (m *mockLCD) Clear() error {
	m.log.Info("Clear")
	return nil
}

func (m *mockLCD) Home() error {
	m.log.Info("Home")
	return nil
}

func (m *mockLCD) SetPosition(line, pos int) error {
	m.log.Infof("SetPosition(%d, %d)", line, pos)
	return nil
}

func (m *mockLCD) Command(cmd byte) error {
	m.log.Infof("Command(%+v)", cmd)
	return nil
}
