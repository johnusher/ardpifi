package gpio

// read switch input from raspberry pi 3+ GPIO input and light LED on GPIO output
// uses command-line GPIOD.
// debouncing handled

// input:  Physical pin 13 = BCM pin 27, GPIO27 = J8p13
// The library uses the raw BCM2835 pin numbers, not the ports as they are mapped
// on the J8 output pins for the Raspberry Pi.
// A mapping from J8 to BCM is provided for those wanting to use the J8 numbering.
// eg physica; pin

import (
	"fmt"
	"math/rand"
	"sync"

	// "log"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/johnusher/ardpifi/pkg/wavs"
	log "github.com/sirupsen/logrus"

	"github.com/warthog618/gpiod"
)

type GPIO interface {
	//Run() error
	Close() error
}

type GPIOMessage struct {
	ButtonFlag     int16
	ButtonDownTime time.Time
}

type gpio struct {
	gpio           chan<- GPIOMessage
	buttonDownTime time.Time
	buttonFlag     int16 // official curent status of button 0= down

	lastEvent      int       // drop duplicate up/up or down/down
	lastDownUpTime time.Time // last time we detected a full down/up

	button          gpiod.Line
	led             gpiod.Line
	buttonWavs      wavs.Wavs
	cancelButtonWav []chan struct{}
	sync.Mutex      // protects cancelButtonWav
}

// type gpio struct {
// 	gpio 	chan<- GPIOMessage
// 	buttonDownTime  time.Time
// 	buttonFlag      int16
// 	cancelButtonWav []chan struct{}
// }

// func Init(gpioChan chan<- GPIOMessage, mock bool) (GPIO, error) {
func Init(gpioChan chan<- GPIOMessage, noSound bool) (GPIO, error) {
	// if noSound { // TBD
	// 	return initMockGPS(gpsChan)
	// }

	return initGPIO(gpioChan, noSound)
}

func initGPIO(gpioChan chan<- GPIOMessage, noSound bool) (GPIO, error) {

	wavsp := wavs.InitWavs(noSound)

	// pushButton is the struct we want to send out
	pushButton := &gpio{
		gpio:            gpioChan,
		buttonFlag:      0,
		buttonDownTime:  time.Now(),
		cancelButtonWav: make([]chan struct{}, 0),
	}
	pushButton.buttonWavs = *wavsp
	buttonEventHandler := mkButtonEventHandler(pushButton)

	// on pi 4, we need to set GPIO 27 to pull up like this:
	// raspi-gpio set 27 pu

	app := "raspi-gpio"
	arg0 := "set"
	arg1 := "27"
	arg2 := "pu"
	cmd := exec.Command(app, arg0, arg1, arg2)
	// log.Printf("gpio set-up part 1")
	err := cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		return nil, err
	}

	// hack from https://www.raspberrypi.org/forums/viewtopic.php?t=270376:

	//  Physical pin 13 = BCM pin 27, GPIO27 = J8p13
	// gpio readall

	// The library uses the raw BCM2835 pin numbers, not the ports as they are mapped
	// on the J8 output pins for the Raspberry Pi.
	// A mapping from J8 to BCM is provided for those wanting to use the J8 numbering.
	// eg physica; pin

	app = "gpio"
	arg0 = "-g"
	arg1 = "mode"
	arg2 = "27"
	arg3 := "in"

	cmd = exec.Command(app, arg0, arg1, arg2, arg3)
	// log.Printf("gpio set-up part 1")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		return nil, err
	}

	arg2 = "22"
	arg3 = "out"

	cmd = exec.Command(app, arg0, arg1, arg2, arg3)
	// log.Printf("gpio set-up part 1.1")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
		return nil, err
	}

	// Set up GPIO 27 with pullup resistor
	app = "gpio"

	arg0 = "-g"
	arg1 = "mode"
	arg2 = "27"
	arg3 = "up"

	cmd = exec.Command(app, arg0, arg1, arg2, arg3)
	// log.Printf("gpio set-up part 2...")
	err = cmd.Run()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	// log.Printf("gpio set-up part 3...")
	c, err := gpiod.NewChip("gpiochip0")
	if err != nil {
		// panic(err)
		return nil, err
	}

	defer c.Close() // should this go here or in the close func??

	// Set up button with interrupt watch using gpiod
	// offset := rpi.J8p13
	offset := 27
	buttonp, err := c.RequestLine(offset,
		gpiod.WithBothEdges,
		gpiod.WithEventHandler(buttonEventHandler))

	if err != nil {
		fmt.Printf("RequestLine returned error: %s\n", err)
		if err == syscall.Errno(22) {
			fmt.Println("Note that the WithPullUp option requires kernel V5.5 or later - check your kernel version.")
			return nil, err
		}
		os.Exit(1)
		return nil, err
	}

	pushButton.button = *buttonp

	// defer pushButton.button.Close()

	// Set up button with interrupt watch using gpiod
	// offset := rpi.J8p13
	offset = 22
	ledp, err := c.RequestLine(offset, gpiod.AsOutput(0)) // during request

	pushButton.led = *ledp

	// NB remove pullup from the gpiod function call: requires kernel 5.5 for pullup/pulldown support.
	if err != nil {
		fmt.Printf("RequestLine2 returned error: %s\n", err)
		os.Exit(1)
		return nil, err
	}
	// defer pushButton.led.Close()  // should this go here or in close func??

	return &gpio{gpioChan,
		time.Now(),               // buttonDownTime
		0,                        // buttonFlag
		-1,                       // lastEvent int
		time.Now(),               // lastDownUpTime time.Time
		pushButton.button,        // button   gpiod.Line
		pushButton.led,           // 	led  gpiod.Line
		pushButton.buttonWavs,    // 	buttonWavs   wavs.Wavs
		make([]chan struct{}, 0), // cancelButtonWav []chan struct{}
		sync.Mutex{},             // sync.Mutex      // protects cancelButtonWav
	}, nil

	// type gpio struct {
	// 	gpio           chan<- GPIOMessage
	// 	buttonDownTime time.Time
	// 	buttonFlag     int16 // official curent status of button 0= down

	// 	lastEvent      int       // drop duplicate up/up or down/down
	// 	lastDownUpTime time.Time // last time we detected a full down/up

	// 	button          gpiod.Line
	// 	led             gpiod.Line
	// 	buttonWavs      wavs.Wavs
	// 	cancelButtonWav []chan struct{}
	// 	sync.Mutex      // protects cancelButtonWav
	// }

}

func (g *gpio) Close() error {
	// c.Close()
	g.button.Close()
	return g.led.Close()
}

// func (g *gpio) Run() error {

// 	for {
// 		g.gpio <- GPIOMessage{

// 		}
// 	}

// }

func delayedButtonHandle(pushButton *gpio) {
	buttonStatus, _ := pushButton.button.Value() // Read state from line (active / inactive)

	pushButton.gpio <- GPIOMessage{int16(buttonStatus), time.Now()}

	if buttonStatus == 0 { // low= button pressed down
		pushButton.led.SetValue(1) // light LED

		// first down: restart timer
		t := time.Now()
		pushButton.buttonDownTime = t

		// play long howl
		catMeowN := rand.Int31n(2) + 1
		catcat := fmt.Sprintf("howl%d.wav", catMeowN)

		cancelButtonWav := pushButton.enqueue()
		go func() {
			// either play after 150ms, or bail if close(cancelButtonWav) is called
			select {
			case <-time.After(150 * time.Millisecond):
				// log.Info("howl!", catcat)
				pushButton.buttonWavs.Play(catcat)

			case <-cancelButtonWav:

			}

		}()

		// wavss.Play("meow_1.wav")

	} else {
		// button has been lifted
		pushButton.led.SetValue(0) // turn off LED
		now := time.Now()
		elapsedTime := now.Sub(pushButton.buttonDownTime)

		pushButton.flush()
		// newtimer2.Stop()
		pushButton.buttonWavs.StopAll()

		if elapsedTime < 400*time.Millisecond {
			// play short bark:
			catMeowN2 := rand.Int31n(9) + 1
			catcat2 := fmt.Sprintf("bark%d.wav", catMeowN2)
			pushButton.buttonWavs.Play(catcat2)
			fmt.Println(elapsedTime)

		}

	}
	pushButton.buttonFlag = 0

}

func mkButtonEventHandler(pushButton *gpio) func(gpiod.LineEvent) {

	return func(evt gpiod.LineEvent) {

		buttonStatus, _ := pushButton.button.Value()
		if buttonStatus == pushButton.lastEvent {
			return
		}

		pushButton.lastEvent = buttonStatus

		// if pushButton.buttonFlag == 0 {

		// 	pushButton.buttonFlag = 1 // flag =1 , ie button active

		//go func() {
		// time.Sleep(8 * time.Millisecond)
		// pushButton.gpio <- GPIOMessage{int16(1), time.Now()}
		delayedButtonHandle(pushButton)
		//}()

		// } else {
		// 	// timer already running
		// 	// log.Info("bounce")
		// 	return
		// }

	}
}

// enqueue creates a new channel for each button pressed down (buttonStatus == 0)
// The new channel is added to the cancelButtonWav slice, and returned to the
// caller. Channel is closed via `flush`.
func (b *gpio) enqueue() chan struct{} {
	b.Lock()
	c := make(chan struct{})
	b.cancelButtonWav = append(b.cancelButtonWav, c)
	b.Unlock()
	return c
}

// flush closes all channels created via `enqueue`, and then clears the slice of
// channels, `cancelButtonWav`.
func (b *gpio) flush() {
	b.Lock()
	for _, c := range b.cancelButtonWav {
		close(c)
	}
	b.cancelButtonWav = make([]chan struct{}, 0)
	b.Unlock()
}
