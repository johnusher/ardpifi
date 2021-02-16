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

	d, err := monochromeoled.Open(&i2c.Devfs{Dev: "/dev/i2c-1"})
	if err != nil {
		panic(err)
	}
	defer d.Close()

	time.Sleep(2 * time.Second)

	// clear the display before putting on anything
	if err := d.Clear(); err != nil {
		panic(err)
	}
	if err := d.SetImage(0, 0, m); err != nil {
		panic(err)
	}
	if err := d.Draw(); err != nil {
		panic(err)
	}
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{200, 100, 0, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}
