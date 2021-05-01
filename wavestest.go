// wavestest.gop

package main

import (
	"time"

	"github.com/johnusher/ardpifi/pkg/wavs"
)

func main() {
	w := wavs.InitWavs()
	w.Play("ceottk001_human.wav")
	time.Sleep(5 * time.Second)
}
