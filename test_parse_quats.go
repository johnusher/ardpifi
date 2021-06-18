package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/image/bmp"
)

const (
	// quat_in     = "./letters/M/M_20-45-12/quaternion_data.txt"
	quat_in = "./letters/O/O_20-32-50/quaternion_data.txt"

	circBufferL = 600 // length of buffer where we store quat data. 600 samples @5 ms update = 3 seconds
	lp          = 28  // pixels used to represent drawn letter, on each axis, ie lpxlp
)

func main() {

	// put following on the stack
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

	f, err := os.Open(quat_in)
	if err != nil {
		fmt.Println(err)
	}

	n = 0 // index to write into circ buffer
	for {

		var flt1, flt2, flt3, flt4 float64
		// var str string

		fn, err := fmt.Fscan(f, &flt1, &flt2, &flt3, &flt4)
		if fn == 0 || err != nil {
			fmt.Printf("err %s\n", err)
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

	startTime := time.Now()

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
	 y_direction[1] = eye_minus_cdscc[1] / norm_y_direction;
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

	minX :=1.0
	maxX :=-1.0
	minY :=minX
	maxY :=maxX
	for i := 0; i < n; i++ {
		// x[i] = x_direction[0]*projected_circ_buffer[i][0] + x_direction[2]*projected_circ_buffer[i][2]                               //  % x_direction(2) is so small we can ignore it
		x[i] = x_direction[0]*projected_circ_buffer[i][0] + x_direction[1]*projected_circ_buffer[i][1] + x_direction[2]*projected_circ_buffer[i][2]      
		y[i] = y_direction[0]*projected_circ_buffer[i][0] + projected_circ_buffer[i][1] + y_direction[2]*projected_circ_buffer[i][2] // ; % y_direction(2) -> 1.0
		minX = math.Min(minX,x[i] )
		maxX = math.Max(maxX,x[i] )
		minY = math.Min(minY,y[i] )
		maxY = math.Max(maxY,y[i] )
	}

        
	// // % scale

	absMixX := math.Abs(minX)
	if(absMixX > maxX){
		maxX = absMixX
	}	


	absMinY:= math.Abs(minY)
	if(absMinY > maxY){
		maxY = absMinY
	}


	maxdim := math.Max(maxX,maxY)     
	scaler := 0.9/maxdim
	// scaler := 1.0/maxdim
	for i := 0; i < n; i++ {
		x[i] = x[i] *scaler
		y[i] = y[i] *scaler

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

		// x_int = math.Max(x_int,lp/2)
		// y_int = math.Max(y_int,lp/2)
		letterImage[y_int][x_int] = 1
	}

	// Create png image
	img := image.NewRGBA(image.Rect(0, 0, lp, lp))

	for i := 0; i < n; i++ {
		x_int = int(x[i]*lp/2 + lp/2 + 1)
		y_int = int(y[i]*lp/2 + lp/2 + 1)
		img.Set(y_int, x_int, color.RGBA{255, 0, 0, 255})
	}

	now1 := time.Now()
	elapsedTime := now1.Sub(startTime)

	// about 200 uS
	log.Printf("elapsedTime1=%v", elapsedTime)

	// Save to out.bmp
	fo, err := os.OpenFile("out5.bmp", os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Printf("err %s\n", err)
	}

	defer fo.Close()
	bmp.Encode(fo, img)

	now2 := time.Now()
	elapsedTime = now2.Sub(now1)

	// takes bout 1.5 ms to save bmp, 10 ms to save png
	log.Printf("elapsedTime2=%v", elapsedTime)

}
