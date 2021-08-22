package main

// read in quaternion data from /letters directory, (iterate over all examples)
// process to a 28x28 array
// convert to base64
// pipe to a tensorflow lite model in python
// receive string response from tflite idicating letter match and probability

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/image/bmp"
)

const (
	circBufferL = 600 // length of buffer where we store quat data. 600 samples @5 ms update = 3 seconds
	lp          = 28  // pixels used to represent drawn letter, on each axis, ie lpxlp
)

func main() {

	var quat_inFn = "./letters/M/M_20-46-09/quaternion_data.txt"

	var quat_in_circ_buffer [circBufferL][5]float64   // raw quaternion inputs from file or IMU
	var projected_circ_buffer [circBufferL][3]float64 // quat projected to xyz
	var centre [3]float64                             // 3x1 averages of projected_circ_buffer coluns
	var centre_direction [3]float64                   // 3x1
	var eye_minus_cdscc [3]float64
	var y_direction [3]float64
	var centre_direction_sq_centre_col [3]float64
	var x_direction [3]float64
	var x [circBufferL]float64
	var y [circBufferL]float64
	// var letterImage [lp][lp]uint8
	var letterImage [lp][lp]byte

	var n int

	searchDir := "letters"

	pattern := "quaternion_data.txt"

	fileList := make([]string, 0)
	err := filepath.Walk(searchDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return err
	})

	if err != nil {
		fmt.Println(err)
	}

	cmd := exec.Command("python3", "-u", "classifier/classify.py") // linux
	if runtime.GOOS == "windows" {
		cmd = exec.Command("python", "-u", "classifier/classify.py") // windoze
	}

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

	for _, file := range fileList {

		matched, _ := regexp.MatchString(pattern, file)
		// fmt.Println(matched, err)

		if matched {
			log.Printf(" file: %v", file)
			quat_inFn = file

			// log.Printf(" fileList: %v", fileList)

			f, err := os.Open(quat_inFn)
			if err != nil {
				fmt.Println(err)
			}

			n = 0 // index to write into circ buffer
			for {

				var flt1, flt2, flt3, flt4 float64

				fn, err := fmt.Fscan(f, &flt1, &flt2, &flt3, &flt4)
				if fn == 0 || err != nil {
					break
				}

				quat_in_circ_buffer[n][0] = flt1
				quat_in_circ_buffer[n][1] = flt2
				quat_in_circ_buffer[n][2] = flt3
				quat_in_circ_buffer[n][3] = flt4

				// log.Printf("quat_in_circ_buffer[n][0] %v", quat_in_circ_buffer[n][0])
				// log.Printf("quat_in_circ_buffer[n][1] %v", quat_in_circ_buffer[n][1])
				// log.Printf("quat_in_circ_buffer[n][2] %v", quat_in_circ_buffer[n][2])
				// log.Printf("quat_in_circ_buffer[n][3] %v", quat_in_circ_buffer[n][3])

				s := quat_in_circ_buffer[n][0]
				x := quat_in_circ_buffer[n][1]
				y := quat_in_circ_buffer[n][2]
				z := quat_in_circ_buffer[n][3]

				projected_circ_buffer[n][0] = 1.0 - 2.0*(y*y+z*z)
				projected_circ_buffer[n][1] = 2.0 * (x*y + s*z)
				projected_circ_buffer[n][2] = 2.0 * (x*z - s*y)

				// log.Printf("projected_circ_buffer[n][0] %v", projected_circ_buffer[n][0])
				// log.Printf("projected_circ_buffer[n][1] %v", projected_circ_buffer[n][1])
				// log.Printf("projected_circ_buffer[n][2] %v", projected_circ_buffer[n][2])

				n = n + 1

			}

			// eof reached

			// startTime := time.Now()

			// step 3. when we have stopped recording data: average of projected

			// centre := mean(projected, 1)

			sumx := 0.0
			sumz := 0.0
			sumy := 0.0

			for i := 0; i < n; i++ {
				sumx += (projected_circ_buffer[i][0])
				sumy += (projected_circ_buffer[i][1])
				sumz += (projected_circ_buffer[i][2])
			}

			centre[0] = (sumx) / (float64(n))
			centre[1] = (sumy) / (float64(n))
			centre[2] = (sumz) / (float64(n))

			// %% step 4: norm-centre

			// %     centre_direction = centre' ./ norm(centre);

			norm_centre := math.Sqrt(centre[0]*centre[0] + centre[1]*centre[1] + centre[2]*centre[2])

			centre_direction[0] = centre[0] / norm_centre
			centre_direction[1] = centre[1] / norm_centre
			centre_direction[2] = centre[2] / norm_centre

			// % step 5: y direction:
			// %     y_direction = (eye[2] - centre_direction*centre_direction')* [0 1 0]';

			centre_direction_sq_centre_col[0] = centre_direction[0] * centre_direction[1]
			centre_direction_sq_centre_col[1] = centre_direction[1] * centre_direction[1]
			centre_direction_sq_centre_col[2] = centre_direction[1] * centre_direction[2]

			eye_minus_cdscc[0] = -1.0 * centre_direction_sq_centre_col[0]
			eye_minus_cdscc[1] = 1.0 - centre_direction_sq_centre_col[1]
			eye_minus_cdscc[2] = -1.0 * centre_direction_sq_centre_col[2]

			// %     y_direction = y_direction ./ norm(y_direction);

			norm_y_direction := math.Sqrt(eye_minus_cdscc[0]*eye_minus_cdscc[0] + eye_minus_cdscc[1]*eye_minus_cdscc[1] + eye_minus_cdscc[2]*eye_minus_cdscc[2])

			y_direction[0] = eye_minus_cdscc[0] / norm_y_direction
			y_direction[1] = eye_minus_cdscc[1] / norm_y_direction
			// y_direction[1] = 1.0 // tends to unity
			y_direction[2] = eye_minus_cdscc[2] / norm_y_direction

			// %%     step 6: x_direction via cross product
			// %     x_direction_cp = cross(centre_direction, y_direction);

			x_direction[0] = centre_direction[1]*y_direction[2] - centre_direction[2]*y_direction[1]
			// %     x_direction[1] = centre_direction[2]*y_direction[0] - centre_direction[0]*y_direction[2]; % very close to zero
			// x_direction(2) = centre_direction(3)*y_direction(1) - centre_direction(1)*y_direction(3); % very close to zero
			x_direction[1] = centre_direction[2]*y_direction[0] - centre_direction[0]*y_direction[2]
			x_direction[2] = centre_direction[0]*y_direction[1] - centre_direction[1]*y_direction[0]

			// %% step 7: x and y corrodinates:
			// x(n) = x_direction(1)* projected(n, 1) + x_direction(2)* projected(n, 2)  + x_direction(3)* projected(n, 3) ;
			// y(n) = y_direction(1)* projected(n, 1) + y_direction(2)* projected(n, 2)  + y_direction(3)* projected(n, 3) ;

			minX := 1.0
			maxX := -1.0
			minY := minX
			maxY := maxX
			for i := 0; i < n; i++ {
				// x[i] = x_direction[0]*projected_circ_buffer[i][0] + x_direction[2]*projected_circ_buffer[i][2]                               //  % x_direction(2) is so small we can ignore it
				x[i] = x_direction[0]*projected_circ_buffer[i][0] + x_direction[1]*projected_circ_buffer[i][1] + x_direction[2]*projected_circ_buffer[i][2]
				y[i] = y_direction[0]*projected_circ_buffer[i][0] + projected_circ_buffer[i][1] + y_direction[2]*projected_circ_buffer[i][2] // ; % y_direction(2) -> 1.0
				minX = math.Min(minX, x[i])
				maxX = math.Max(maxX, x[i])
				minY = math.Min(minY, y[i])
				maxY = math.Max(maxY, y[i])
			}

			// // % scale

			absMixX := math.Abs(minX)
			if absMixX > maxX {
				maxX = absMixX
			}

			absMinY := math.Abs(minY)
			if absMinY > maxY {
				maxY = absMinY
			}

			maxdim := math.Max(maxX, maxY)
			scaler := 0.8 / maxdim // scale image so we don't extend to the edge: this REALLY help %prob!

			for i := 0; i < n; i++ {
				x[i] = x[i] * scaler
				y[i] = y[i] * scaler

			}

			// make black and white image:

			// first blank the image:
			for ix := 0; ix < lp; ix++ {
				for iy := 0; iy < lp; iy++ {
					letterImage[ix][iy] = 0
				}
			}

			x_int := 0
			y_int := 0
			for i := 0; i < n; i++ {
				x_int = int(x[i]*lp/2 + lp/2)
				y_int = int(y[i]*lp/2 + lp/2)
				letterImage[y_int][x_int] = 1
			}

			var joinedArray []byte

			// resize matrix into long array:
			for ix := 0; ix < lp; ix++ {
				nums := letterImage[ix][:]
				joinedArray = append(joinedArray, nums...)
				// joinedArray := bytes.Join(nums, nil)  // dunno how to do this
			}

			encoded := base64.StdEncoding.EncodeToString(joinedArray)

			// now send to the python:

			now1 := time.Now()

			// send encoded base64 28x28 ti TF:
			_, err = stdin.Write([]byte(encoded))
			if err != nil {
				log.Errorf("stdin.Write() failed: %s", err)
			}

			// write end of line:
			_, err = stdin.Write([]byte("\n"))
			if err != nil {
				log.Errorf("stdin.Write() failed: %s", err)
			}

			s2, err := stdoutReader.ReadString('\n')
			if err != nil {
				log.Printf("Process is finished ..")
			}

			now2 := time.Now()
			elapsedTime := now2.Sub(now1)
			log.Printf("elapsedTime TF=%v", elapsedTime)

			log.Printf("raw message: %v", s2)

			// // print first and second place:
			// s := strings.FieldsFunc(s2, Split)

			// prob, _ := strconv.ParseFloat(s[0], 64)
			// // letter := strings.Trim(s[1], "'")
			// letter := strings.Replace(s[1], "'", "", -1)

			// // s[2] is blank
			// prob2, _ := strconv.ParseFloat(s[3], 64)
			// letter2 := strings.Replace(s[4], "'", "", -1)

			// log.Printf("letter1: %v", letter)
			// log.Printf("prob1: %v", prob)

			// log.Printf("letter2: %v", letter2)
			// log.Printf("prob2: %v", prob2)

			// -------------------------------
			// Create png image
			img := image.NewRGBA(image.Rect(0, 0, lp, lp))

			for i := 0; i < n; i++ {
				x_int = int(x[i]*lp/2 + lp/2 + 1)
				y_int = int(y[i]*lp/2 + lp/2 + 1)
				img.Set(y_int, x_int, color.RGBA{255, 0, 0, 255})
			}

			// now1 = time.Now()
			// elapsedTime = now1.Sub(startTime)

			// about 200 uS
			// log.Printf("elapsedTime1=%v", elapsedTime)

			// Save to out.bmp

			sn := strings.Replace(file, "quaternion_data.txt", "quat_image.bmp", 1)

			fo, err := os.OpenFile(sn, os.O_WRONLY|os.O_CREATE, 0600)
			if err != nil {
				log.Printf("err %s\n", err)
			}

			defer fo.Close()
			bmp.Encode(fo, img)

			// now2 := time.Now()
			// elapsedTime = now2.Sub(now1)

			// takes bout 1.5 ms to save bmp, 10 ms to save png
			// log.Printf("elapsedTime2=%v", elapsedTime)

		}

	}

}

// func request(r *bufio.Reader, w io.Writer, str string) string {
// 	w.Write([]byte(str))
// 	w.Write([]byte("\n"))
// 	str, err := r.ReadString('\n')
// 	if err != nil {
// 		panic(err)
// 	}
// 	return str[:len(str)-1]
// }

func copyOutput(r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
}

func ReadOutput(rc io.ReadCloser) (string, error) {
	x, err := ioutil.ReadAll(rc)
	s := string(x)
	return s, err
}

func ReadOutput2(output chan string, rc io.ReadCloser) {
	r := bufio.NewReader(rc)
	for {
		x, _ := r.ReadString('\n')
		output <- string(x)
	}
}

func sliceToInt(s []byte) byte {
	res := int(0)
	op := int(1)
	for i := len(s) - 1; i >= 0; i-- {
		res += int(s[i]) * op
		op *= 2

		// log.Printf(" res: %v", res)

	}
	return byte(res)
}

func Split(r rune) bool {
	return r == ':' || r == ',' || r == '(' || r == ')' || r == '[' || r == ']'
}
