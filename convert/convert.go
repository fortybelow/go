package convert

import( "fmt"; "io"; "os"; "io/ioutil" )

var out io.Writer;


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

type Hex_t    []uint8 // each uint8 contains 4 bits
type Base64_t []uint8 // each uint8 contains 6 bits

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
	x := make(Hex_t, len(n))

	for i := 0; i < len(n); i++ {
		x[i] = m[ n[i] ]
	}
	return string(x)
}

func HexToValue(s string) Hex_t {
	m := Map_h()
	x := make(Hex_t, len(s))

	for i := 0; i < len(s); i++ {
		x[i] = m[ s[i] ]
	}
	return x
}


func printMap_k(m map[uint8]uint8) {
	for k, v := range m {
		fmt.Fprintf(os.Stdout, "%v : %v\n", string(k), v)
	}
}

func printMap_v(m map[uint8]uint8) {
	for k, v := range m {
		fmt.Fprintf(os.Stdout, "%v : %v\n", k, string(v))
	}
}


// Expects hex in Big Endian order
// TODO{ modify s in place and return truncated s}
func HexToBase64(s Hex_t) Base64_t {
	rep  := make([]uint8, len(s))
	idx  := len(rep) - 1

	for bit  := 0; bit + 6 <= 4 * len(s); bit += 6 {
		firstByte  := s[len(s) - 2 - bit / 4]
		secondByte := s[len(s) - 1 - bit / 4]

		if bit % 4 == 0 {
			l, r := firstByte & 0x3, secondByte
			rep[idx] = l << 4 + r
		} else {
			l, r := firstByte, secondByte & 0xC >> 2
			rep[idx] = l << 2 + r
		}

		idx -= 1
	}

	switch 4 * len(s) % 6 {
		case 2:
			rep[idx] = s[0] & 0xC >> 2
			idx -= 1
		case 4:
			rep[idx] = s[0]
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
		x[k] = (lhs[len(lhs) - i] ^ rhs[len(rhs) - j]) & 0xF
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
	debugModeActivated := true

	if debugModeActivated {
	    out = os.Stdout
	} else {
		out = ioutil.Discard
	}

	if len(os.Args) != 2 && len(os.Args) != 3 {
		fmt.Printf("Usage: ./a <Hex string> [<Hex string>]\n")
		return
	} else if len(os.Args) == 2 {
		fmt.Println(HexToBase64(HexToValue(os.Args[1])))
	} else {
		x1 := HexToValue(os.Args[1])
		x2 := HexToValue(os.Args[2])
		fmt.Println(x1)
		fmt.Println(x2, "^")

		x3 := Xor_h(x1, x2)
		fmt.Println(x3)
		
	}
}