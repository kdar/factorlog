package factorlog

import (
	"runtime"
)

// stacks returns a stack trace from the runtime
// if all is true, all goroutines are included
func stacks(all bool) []byte {
	n := 10000
	if all {
		n = 100000
	}
	var trace []byte
	for i := 0; i < 5; i++ {
		trace = make([]byte, n)
		nbytes := runtime.Stack(trace, all)
		if nbytes < len(trace) {
			return trace[:nbytes]
		}
		n *= 2
	}
	return trace
}

const digits = "0123456789"

// twoDigits converts an integer d to its ascii representation
// i is the destination index in buf
func twoDigits(buf *[]byte, i, d int) {
	(*buf)[i+1] = digits[d%10]
	d /= 10
	(*buf)[i] = digits[d%10]
}

// nDigits converts an integer d to its ascii representation
// n is how many digits to use
// i is the destination index in buf
func nDigits(buf *[]byte, n, i, d int) {
	// reverse order
	for j := n - 1; j >= 0; j-- {
		(*buf)[i+j] = digits[d%10]
		d /= 10
	}
}

const ddigits = `0001020304050607080910111213141516171819` +
	`2021222324252627282930313233343536373839` +
	`4041424344454647484950515253545556575859` +
	`6061626364656667686970717273747576777879` +
	`8081828384858687888990919293949596979899`

// itoa converts an integer d to its ascii representation
// i is the deintation index in buf
// algorithm from https://www.facebook.com/notes/facebook-engineering/three-optimization-tips-for-c/10151361643253920
func itoa(buf *[]byte, i, d int) int {
	j := len(*buf)

	for d >= 100 {
		// Integer division is slow, so we do it by 2
		index := (d % 100) * 2
		d /= 100
		j--
		(*buf)[j] = ddigits[index+1]
		j--
		(*buf)[j] = ddigits[index]
	}

	if d < 10 {
		j--
		(*buf)[j] = byte(int('0') + d)
		return copy((*buf)[i:], (*buf)[j:])
	}

	index := d * 2
	j--
	(*buf)[j] = ddigits[index+1]
	j--
	(*buf)[j] = ddigits[index]

	return copy((*buf)[i:], (*buf)[j:])
}
