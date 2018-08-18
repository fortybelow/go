package main

import(
	"fmt"
	cv "github.com/fortybelow/go/convert"
)

func Freq_h(h cv.Hex_t) map[uint8]int {
	f := make(map[uint8]int)
	for _, v := range h {
		f[v] += 1
	}
	return f
}


func Digits_h() func() uint8 {
	var l, r uint8 = 0, 0
	return func() uint8 {
		x, y := l, r

		if r == 15 {
			if l != 15 {
				l += 1
				r = 0
			}
		} else {
			r += 1
		}
		return x << 4 + y
	}
}

func isValid(ch uint8) bool {
	return ch >= uint8('a') && ch <= uint8('z') ||
		   ch >= uint8('A') && ch <= uint8('Z') ||
		   ch == uint8(' ')
}

func Count(m map[uint8]int) int {
	count := 0

	for k, v := range m {
		if isValid(k) {
			count += v
		}
	}

	return count
}

func main() {
	h   := cv.Decode_h("1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736")
	gen := Digits_h()

	var key, message cv.Hex_t
	maxCount := 0

	for i := 0; i < 256; i++ {
		x := make(cv.Hex_t, len(h))
		p := gen()
		
		for j := 0; j < len(x); j++ {
			x[j] = p
		}

		y := cv.Xor_h(h, x)
		f := Freq_h(y)

		if c := Count(f); c > maxCount {
			maxCount  = c
			key       = x
			message   = y
		}
	}

	fmt.Printf("Count: %v  Key: %v\n", maxCount, key)
	fmt.Printf("%v\n", cv.Cast_htos(message))
}