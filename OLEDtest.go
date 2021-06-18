package main

import (
	"image"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

	"image/color"
	_ "image/png"

	"time"

	"github.com/goiot/devices/monochromeoled"
	"golang.org/x/exp/io/i2c"
)

func main() {

	img := image.NewRGBA(image.Rect(0, 0, 128, 64))

	addLabel(img, 20, 30, "Maxii!!")
	addLabel(img, 20, 60, "daddy!!")

	// f, err := os.Create("out.png")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()
	// if err := png.Encode(f, img); err != nil {
	// 	panic(err)
	// }

	// rc, err := os.Open("./maxi.png")

	// if err != nil {
	// 	panic(err)
	// }
	// defer rc.Close()

	// m, _, err := image.Decode(rc)
	// if err != nil {
	// 	panic(err)
	// }

	m := img

	oled, err := monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err != nil {
		panic(err)
	}
	defer oled.Close()

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

	for {
		t := time.Now()
		// lcd.SetPosition(1, 0)
		// fmt.Fprint(lcd, t.Format("Monday Jan 2"))
		// lcd.SetPosition(2, 1)
		// fmt.Fprint(lcd, t.Format("15:04:05 2006"))

		// if err := oled.Clear(); err != nil {
		// 	panic(err)
		// }

		clearLine(img, 1)
		addLabel(img, 0, 1, t.Format("15:04:05 2006"))

		if err := oled.SetImage(0, 0, img); err != nil {
			panic(err)
		}
		if err := oled.Draw(); err != nil {
			panic(err)
		}

		time.Sleep(666 * time.Millisecond)

	}

}

func addLabel(img *image.RGBA, x, line int, label string) {

	lineOffset := (line) * 10
	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(lineOffset * 64)}

	oled := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	oled.DrawString(label)
}

func clearLine(img *image.RGBA, line int) {
	col := color.RGBA{0, 0, 0, 0}
	// w := oled.Width()
	w := 128 // pixel width of OLED screen
	lineOffset := (line - 1) * 10
	for y := 1; y < 10; y++ {
		for x := 1; x < w; x++ {
			img.Set(x, y+lineOffset, col)
		}
	}

}
