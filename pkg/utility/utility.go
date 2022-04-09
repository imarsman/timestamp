package utility

import "strings"

// DigitCount count digits in an int64 number
func DigitCount(number int64) int64 {
	var count int64 = 0
	for number != 0 {
		number /= 10
		count++
	}
	return count
}

// BytesToString convert byte list to string with no allocation
//
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func BytesToString(bytes ...byte) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(bytes); i++ {
		sb.WriteByte(bytes[i])
	}
	return sb.String()
}

// RunesToString convert runes list to string with no allocation
//
// WriteRune is more complex than WriteByte so can't inline
//
// A small cost a few ns in testing is incurred for using a string builder.
// There are no heap allocations using strings.Builder.
func RunesToString(runes ...rune) string {
	var sb = new(strings.Builder)
	for i := 0; i < len(runes); i++ {
		sb.WriteRune(runes[i])
	}
	return sb.String()
}

// DaysBefore[m] counts the number of days in a non-leap year
// before month m begins. There is an entry for m=12, counting
// the number of days before January of next year (365).
var DaysBefore = [...]int32{
	0,
	31,
	31 + 28,
	31 + 28 + 31,
	31 + 28 + 31 + 30,
	31 + 28 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30,
	31 + 28 + 31 + 30 + 31 + 30 + 31 + 31 + 30 + 31 + 30 + 31,
}

// 	// Add in days before this month.
// 	d += uint64(daysBefore[month-1])
// 	if isLeap(year) && month >= March {
// 		d++ // February 29
// 	}

// */

// Norm returns nhi, nlo such that
//	hi * base + lo == nhi * base + nlo
//	0 <= nlo < base
// From Go time package
// Example
// Normalize month, overflowing into year.
// m := int(month) - 1
// year, m = Norm(year, m, 12)
// month = Month(m) + 1
func Norm(hi, lo, base int64) (nhi, nlo int64) {
	if lo < 0 {
		n := (-lo-1)/base + 1
		hi -= n
		lo += n * base
	}
	if lo >= base {
		n := lo / base
		hi += n
		lo -= n * base
	}
	return hi, lo
}
