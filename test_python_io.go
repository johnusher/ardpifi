package main

// convert uint8 to base64
// pipe to  python
// receive string response from python

// what we want to do is to keep the pythn app running and just send and receive new data!!

import (
	"encoding/base64"
	"io"
	"io/ioutil"
	"math/rand"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {

	var r int
	var in interface{}
	rand.Seed(time.Now().UnixNano())

	// cmd := exec.Command("python3", "test_go_py_io.py") // linux
	cmd := exec.Command("python", "test_go_py_io.py") // windy

	// // use -u flag if we want unbuffered:
	// // https://stackoverflow.com/questions/55312593/golang-os-exec-flushing-stdin-without-closing-it

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	// we want to start the loop here!!

	r = rand.Intn(8)
	log.Printf("r: %v", r)

	r2 := uint8(r)

	in = []uint8{r2}

	var buf = make([]byte, 1)
	buf = in.([]byte)

	log.Printf("in message: %v", buf)

	encoded := base64.StdEncoding.EncodeToString(buf)

	log.Printf("encoded message: %v", encoded)

	// now send to the python:

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, encoded)
		// io.WriteString(stdin, "\n")
		// stdin.Close()
	}()

	s2, err := ReadOutput(stdout)
	if err != nil {
		log.Printf("Process is finished ..")
	}

	log.Printf("raw message: %v", s2)

	// end loop here!

}

func ReadOutput(rc io.ReadCloser) (string, error) {
	x, err := ioutil.ReadAll(rc)
	s := string(x)
	return s, err
}
