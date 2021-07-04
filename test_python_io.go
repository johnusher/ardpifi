package main

// convert uint8 to base64
// pipe to  python
// receive string response from python

// what we want to do is to keep the pythn app running and just send and receive new data!!

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"math/rand"
	"os/exec"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// cmd := exec.Command("python3", "test_go_py_io.py") // linux
	// cmd := exec.Command("python3", "-u", "test_go_py_io.py") // linux
	cmd := exec.Command("python", "-u", "test_go_py_io.py") // windy -> can we do "python3 -u test_go_py_io.py" on windy?
	// cmd := exec.Command("./test_bash_io.sh")

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

	stdoutReader := bufio.NewReader(stdout)

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	// we want to start the loop here!!

	for {
		r := rand.Intn(8)
		log.Printf("r: %v", r)

		buf := []byte(fmt.Sprintf("%d", r))

		log.Printf("in message: %v", buf)

		encoded := base64.StdEncoding.EncodeToString(buf)

		log.Printf("encoded message: %v", encoded)

		// now send to the python:

		_, err := stdin.Write([]byte(encoded))
		if err != nil {
			log.Errorf("stdin.Write() failed: %s", err)
		}
		_, err = stdin.Write([]byte("\n"))
		if err != nil {
			log.Errorf("stdin.Write() failed: %s", err)
		}

		s2, err := stdoutReader.ReadString('\n')
		if err != nil {
			log.Printf("Process is finished ..")
		}

		log.Printf("raw message: %v", string(s2))
	}

	// end loop here!
}
