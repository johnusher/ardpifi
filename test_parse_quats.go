package main

import (
	"fmt"
	"os"
)

var quat_in = "./letters/M/M_20-45-12/quaternion_data.txt"

func main() {
	f, err := os.Open(quat_in)
	if err != nil {
		fmt.Printf("RequestLine2 returned error: %s\n", err)
		fmt.Println(err)
	}
	for {
		// var i int
		var flt1, flt2, flt3 float64
		// var str string
		var n int
		n, err = fmt.Fscan(f, &flt1, &flt2, &flt3)
		if n == 0 || err != nil {
			fmt.Printf("err %s\n", err)
			break
		}
		fmt.Println("float:", flt1, "; float:", flt2, "; float:", flt3)

		sum := flt1 + flt2

		fmt.Println("sum=", sum)

	}
}
