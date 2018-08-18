package main

import( "fmt"; "io"; "os"; "io/ioutil" )

var out io.Writer = ioutil.Discard;


type InvalidHexCharacter   struct { arg uint8 }
type InvalidBase64Argument struct { arg uint8 }

func (e InvalidHexCharacter)   Error() string { return fmt.Sprintf("Invalid Hex Character Received: %s",    string(e.arg)) }
func (e InvalidBase64Argument) Error() string { return fmt.Sprintf("Invalid Base 64 Argument Received: %s", string(e.arg)) }

// Given a hexadecimal character value, returns the actual value and vice-versa
func Map_h() map[uint8]uint8 {
	m := make(map[uint8]uint8)
	i := uint8(0)

	for ; i < 10; i ++ {
		m['0' + i] = i
		m[i] = '0' + i
	}

	for ; i < 16; i ++ {
		m['A' + (i - 10)] = i
		m['a' + (i - 10)] = i
		m[i] = 'a' + (i - 10)
	}

	return m
}

// Given a base64 value, returns the character value
func Map_b64tov() map[uint8]uint8 {
	m := make(map[uint8]uint8)

	for i := 'A'; i != '9'; i++ {
		m[uint8(len(m))] = uint8(i)

		if i == 'Z' {
			i = 'a' - 1
		} else if i == 'z' {
			i = '0' - 1
		}
	}

	m[uint8(len(m))] = '9'
	m[uint8(len(m))] = '+'
	m[uint8(len(m))] = '/'

	return m
}

// Given a char value, returns the base 64 value
func Map_vtob64() map[uint8]uint8 {
	m := make(map[uint8]uint8)

	for i := 'A'; i != '9'; i++ {
		m[uint8(i)] = uint8(len(m))

		if i == 'Z' {
			i = 'a' - 1
		} else if i == 'z' {
			i = '0' - 1
		}
	}

	m[uint8('9')] = uint8(len(m))
	m[uint8('+')] = uint8(len(m))
	m[uint8('/')] = uint8(len(m))

	return m
}

type Hex_t    []uint8 // each uint8 contains two hex digits
type Base64_t []uint8 // each uint8 contains 6 valid bits

func (n Base64_t) String() string {
	m := Map_b64tov()
	x := make(Base64_t, len(n))

	for i := 0; i < len(n); i++ {
		x[i] = m[ n[i] ]
	}
	return string(x)
}

func (n Hex_t) String() string {
	m := Map_h()
	x := make(Hex_t, len(n) * 2)

	for i := 0; i < len(n); i++ {
		x[2 * i    ] = m[ n[i] & 0xF ]
		x[2 * i + 1] = m[ n[i] & 0xF0 >> 4 ]
	}

	return string(x)
}


func Decode_h(s string) Hex_t {
	m := Map_h()
	b := len(s) % 2
	x := make(Hex_t, len(s) / 2 + b)

	if b == 1 {
		x[0] = m[s[0]]
		s    = s[1:]
	}

	for i := 0; i < len(s) - 1; i += 2 {
		// Hi << 4 + Lo : Hi Hi Hi Hi Lo Lo Lo Lo
		x[i / 2 + b] = m[s[i]] << 4 + m[s[i + 1]]
	}

	return x
}

func Decode_b64(s string) Base64_t {
	m := Map_vtob64()
	x := make(Base64_t, len(s))

	for i := 0; i < len(s); i++ {
		x[i] = m[ s[i] ]
	}

	return x
}

func Cast_htoc(h Hex_t) string {
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
		fmt.Println("Hex   : ", os.Args[1])
		fmt.Println("Base64: ", Cast_htob64(Decode_h(os.Args[1])))
		fmt.Println("Base64:  SSdtIGtpbGxpbmcgeW91ciBicmFpbiBsaWtlIGEgcG9pc29ub3VzIG11c2hyb29t")

		// fmt.Println("Base64: ", []byte(Cast_htob64(Decode_h(os.Args[1]))))
		// fmt.Println("Base64: ", []byte(Decode_b64("Fc2hyb29t")))
	} else {
		x1 := Decode_h(os.Args[1])
		x2 := Decode_h(os.Args[2])
		fmt.Println(x1)
		fmt.Println(x2, "^")

		x3 := Xor_h(x1, x2)
		fmt.Println(x3)
	}

	fmt.Println(Cast_htoc(Decode_h("48656c6c6f")))
}