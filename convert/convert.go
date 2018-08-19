package convert

import( "fmt"; "io"; "os"; "io/ioutil" )

var out io.Writer = ioutil.Discard;


type InvalidHexCharacter   struct { arg uint8 }
type InvalidBase64Argument struct { arg uint8 }

func (e InvalidHexCharacter)   Error() string { return fmt.Sprintf("Invalid Hex Character Received: %s",    string(e.arg)) }
func (e InvalidBase64Argument) Error() string { return fmt.Sprintf("Invalid Base 64 Argument Received: %s", string(e.arg)) }


// Given a hexadecimal character value, returns the actual value
func Cast_rtoh(ch uint8) uint8 {
	switch {
		case ch >= 'a' && ch <= 'z':
			return ch - 'a' + 10
		case ch >= 'A' && ch <= 'Z':
			return ch - 'A' + 10
		case ch >= '0' && ch <= '9':
			return ch - '0'
	}

	return ch
}

// Given a hexadecimal value, returns the character value
func Cast_htor(b uint8) uint8 {
	switch {
		case b < 10:
			return '0' + b
		default:
			return 'a' + (b - 10)
	}
}

// Given a hexadecimal character value, returns the actual value
func Cast_rtob64(ch uint8) uint8 {
	switch {
		case ch >= 'A' && ch <= 'Z':
			return ch - 'A'
		case ch >= 'a' && ch <= 'z':
			return ch - 'a' + 26
		case ch >= '0' && ch <= '9':
			return ch - '0' + 52
		case ch == '+':
			return 62
		case ch == '/':
			return 63
	}

	return 0
}

func Cast_b64tor(b uint8) uint8 {
	switch {
		case b < 26:
			return 'A' + b
		case b < 52:
			return 'a' + (b - 26)
		case b < 62:
			return '0' + (b - 52)
		case b == 62:
			return '+'
		case b == 63:
			return '/'
	}

	return 0
}


type Hex_t    []uint8 // each uint8 contains two hex digits
type Base64_t []uint8 // each uint8 contains 6 valid bits


func (n Base64_t) String() string {
	x := make(Base64_t, len(n))

	for i := 0; i < len(n); i++ {
		x[i] = Cast_b64tor( n[i] )
	}
	return string(x)
}


func (n Hex_t) String() string {
	x := make(Hex_t, len(n) * 2)

	for i := 0; i < len(n); i++ {
		x[2 * i    ] = Cast_htor( n[i] & 0xF0 >> 4 )
		x[2 * i + 1] = Cast_htor( n[i] & 0x0F )
	}

	return string(x)
}

func Decode_h(s string) Hex_t {
	b := len(s) % 2
	x := make(Hex_t, len(s) / 2 + b)

	if b == 1 {
		x[0] = Cast_rtoh(s[0])
		s    = s[1:]
	}

	for i := 0; i < len(s) - 1; i += 2 {
		// Hi << 4 + Lo : Hi Hi Hi Hi Lo Lo Lo Lo
		x[i / 2 + b] = Cast_rtoh(s[i]) << 4 + Cast_rtoh(s[i + 1])
	}

	return x
}

func Decode_b64(s string) Base64_t {
	x := make(Base64_t, len(s))

	for i := 0; i < len(s); i++ {
		x[i] = Cast_rtob64( s[i] )
	}

	return x
}

func Cast_htos(h Hex_t) string {
	return string(h)
}

// Expects hex in Big Endian order
func Cast_htob64(h Hex_t) Base64_t {
	rep  := make([]uint8, len(h) * 2)
	byt, bit, idx := 0, 0, len(rep) - 1

	for ; bit + 6 < 8 * len(h); bit += 6 {

		switch bit % 24 {
			case 0:
				x   := h[len(h) - 1 - byt]
				byt -= 1
				rep[idx] = x & 0x3F
			case 6:
				x, y := h[len(h) - 1 - byt], h[len(h) - 2 - byt]
				rep[idx] = x & 0xC0 >> 6 + y & 0xF << 2
			case 12:
				x, y := h[len(h) - 1 - byt], h[len(h) - 2 - byt]
				rep[idx] = x & 0xF0 >> 4 + y & 0x03 << 4
			case 18:
				x := h[len(h) - 1 - byt]
				rep[idx] = x & 0xFC >> 2
		}

		byt, idx = byt + 1, idx - 1
	}

	var val uint8

	switch 8 * len(h) - bit {
		case 2:
			val = h[0] & 0xC0 >> 6
		case 4:
			val = h[0] & 0xF0 >> 4
		case 6:
			val = h[0] & 0xFC >> 2
	}

	if val != 0 {
		rep[idx] = val
		idx -= 1
	}

	return rep[idx + 1 :]
}


// Expects Base64 in Big Endian order
func Cast_b64toh(b Base64_t) Hex_t {
	h         := make(Hex_t, len(b))
	bit, hIdx := 0, 0
	var hi, lo, val uint8

	for bIdx := 0; bit + 8 <= 6 * len(b); bit += 8 {
		switch bit % 24 {
			case 0:
				if bit > 0 { bIdx += 1; }

				hi = b[bIdx    ] & 0x3F << 2
				lo = b[bIdx + 1] & 0x30 >> 4
			case 8:
				hi = b[bIdx    ] & 0x0F << 4
				lo = b[bIdx + 1] & 0x3C >> 2
			case 16:
				hi = b[bIdx    ] & 0x03 << 6
				lo = b[bIdx + 1] & 0x3F
		}

		bIdx += 1
		h[hIdx] = hi + lo
		hIdx += 1
	}

	switch 6 * len(b) - bit {
		case 2:
			val = b[0] & 0x30 >> 4
		case 4:
			val = b[0] & 0x3C >> 2
		case 6:
			val = b[0] & 0x3F
	}

	if val != 0 {
		h[hIdx] = val
		hIdx   += 1
	}

	return h[: hIdx]
}

func CountBits(b uint8) int {
	count := 0

	for b > 0 {
		b &= (b - 1)
		count += 1
	}

	return count
}

func HammingDistanceS(a, b string) int {
	distance := 0
	for i := 0; i < len(a); i++ {
		distance += CountBits(a[i] ^ b[i])
	}
	return distance
}

func HammingDistanceR(a, b []uint8) int {
	distance := 0
	for i := 0; i < len(a); i++ {
		distance += CountBits(a[i] ^ b[i])
	}
	return distance
}

func Xor_h(lhs Hex_t, rhs Hex_t) Hex_t {
	max  := func(a int, b int) int { if a > b { return a; } else { return b; } }
	x    := make(Hex_t, max(len(lhs), len(rhs)))
	k    := len(x) - 1
	i, j := 1, 1

	for ; len(lhs) - i >= 0 && len(rhs) - j >= 0; i, j = i + 1, j + 1 {
		x[k] = lhs[len(lhs) - i] ^ rhs[len(rhs) - j]
		k -= 1
	}

	for ; len(lhs) - i >= 0; i++ {
		x[k] = lhs[len(lhs) - i]
		k -= 1
	}

	for ; len(rhs) - j >= 0; j++ {
		x[k] = rhs[len(rhs) - j]
		k -= 1
	}

	return x
}

// 49276d206b696c6c696e6720796f757220627261696e206c696b65206120706f69736f6e6f7573206d757368726f6f6d
// SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t

func main() {
	debugModeActivated := false

	if debugModeActivated {
	    out = os.Stdout
	}

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Printf("Usage: ./a <Hex string> [<Hex string>]\n")
		return
	} else if len(os.Args) == 2 {
		fmt.Println("Hex   : ", Decode_h(os.Args[1]))
		fmt.Println("Base64: ", (Cast_htob64(Decode_h(os.Args[1]))))
		fmt.Println("Should:  SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t")
		fmt.Println("      : ", (Decode_b64("SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t")))

		fmt.Println("Base64: ", []byte(Cast_htob64(Decode_h(os.Args[1]))))
		fmt.Println("Base64: ", []byte(Decode_b64("Fc2hyb29t")))
	} else {
		x1 := Decode_h(os.Args[1])
		x2 := Decode_h(os.Args[2])
		fmt.Println(x1)
		fmt.Println(x2, "^")

		x3 := Xor_h(x1, x2)
		fmt.Println(x3)
	}

	fmt.Println(Cast_htos(Decode_h("48656c6c6f")))
}