package timestamp

import (
	"errors"
	"net/http"
	"time"

	"github.com/JohnCGriffin/overflow"
	"github.com/imarsman/timestamp/pkg/utility"
	"github.com/imarsman/timestamp/pkg/xfmt"
	// gocache "github.com/patrickmn/go-cache"
	// https://golang.org/pkg/time/tzdata/
	/*
		    Package tzdata provides an embedded copy of the timezone database.
		    If this package is imported anywhere in the program, then if the
		    time package cannot find tzdata files on the system, it will use
		    this embedded information.

			Importing this package will increase the size of a program by about
			450 KB.

			This package should normally be imported by a program's main
			package, not by a library. Libraries normally shouldn't decide
			whether to include the timezone database in a program.

			This package will be automatically imported if you build with
			  -tags timetzdata
	*/// This will explicitly include tzdata in a build. See above for build flag.
	// You can do this in the main package if you choose.
	// _ "time/tzdata"
)

// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go/32620397#32620397

// MaxTimestamp max time value
var MaxTimestamp = time.Unix(1<<63-62135596801, 999999999)

// MinTimestamp the minimum timestamp
var MinTimestamp = time.Time{}

// YearDiffOverflows do to year values summed exceed the maximum year value
// Subtractions both ways are tried
func YearDiffOverflows(startYear int64, endYear int64) bool {
	if startYear > endYear {
		return (startYear - endYear) > MaxTimestamp.Unix()
	}
	return (endYear - startYear) > MaxTimestamp.Unix()
}

// YearIsOutOfBounds is year greater than max year or less than min year
func YearIsOutOfBounds(year int64) bool {
	return year > MaxTimestamp.Unix() || year < MinTimestamp.Unix()
}

// TimeIsOutOfBounds is time less than min or greater than max
func TimeIsOutOfBounds(t time.Time) bool {
	return t.Unix() < 0 || int64(t.Year()) > MaxTimestamp.Unix()
}

// YearIsBeyondMax incoming year is greater than the max year
func YearIsBeyondMax(year int64) bool {
	return year > MaxTimestamp.Unix()
}

// YearIsBeyondMin incoming year is less than the min year
func YearIsBeyondMin(year int64) bool {
	return year >= MinTimestamp.Unix()
}

// Int64Overflows does a list of int64s overflow int64?
func Int64Overflows(int64s ...int64) (sum int64, ok bool) {
	for i := 0; i < len(int64s); i++ {
		sum, ok = overflow.Add64(sum, int64s[i])
		if ok == false {
			return sum, true
		}
	}

	return
}

// DurationOverflows does a list of durations overflow int64?
func DurationOverflows(durations ...time.Duration) (sum int64, ok bool) {
	for i := 0; i < len(durations); i++ {
		sum, ok = overflow.Add64(sum, int64(durations[i]))
		if ok == false {
			return sum, true
		}
	}

	return
}

func init() {
}

// Can view allocation analysis with
//   go build -gcflags '-m -m' timestamp.go 2>&1 |less

// OffsetForLocation get offset data for a named zone such a America/Tornto or EST or MST. Based on date the offset for
// a zone can differ, with, for example, an offset of -0500 for EST in the summer and -0400 for EST in the winter. This
// assumes that a year, month, and day is available and have been used to create the date to be analyzed. Based on this
// the offset for the supplied zone name is obtained. This has to be tested more, in particular the calculations to get
// the minutes.
//
// Get integer value of hours offset
//   hours = int(d.Hours())
//
// For 5.5 hours of offset or 0530
//  60 × 5.5 = 330 minutes total offset
//  330 % 60 = 30 minutes
//
// For an offset of 4.25 hours or 0415
//  60 × 4.25 = 255 minutes total offset
//  255 % 60 = 15 minutes
//
// If the zone is not recognized in Go's tzdata database an error will be
// returned.
func OffsetForLocation(year int, month time.Month, day int, locationName string) (duration time.Duration, err error) {
	l, err := time.LoadLocation(locationName)
	if err != nil {
		return 0, err
	}

	t := time.Date(year, month, day, 0, 0, 0, 0, l)
	duration = OffsetForTime(t)

	return
}

// OffsetForTime the duration of the offset from UTC. Mostly the same as doing
// the same thing inline but this reliably gets a duration.
func OffsetForTime(t time.Time) (duration time.Duration) {
	_, offset := t.Zone()

	duration = time.Duration(offset) * time.Second

	return
}

// ZoneFromHM get fixed zone from hour and minute offset
// A negative offsetH will result in a negative zone offset
func ZoneFromHM(offsetH, offsetM int) (location *time.Location) {
	if offsetM < 0 {
		offsetM = -offsetM
	}

	// Must be passed a value equivalent to total seconds for hours and minutes
	location = LocationFromOffset(offsetH*60*60 + offsetM*60)

	return
}

// OffsetHM get hours and minutes for location offset from UTC
// Avoiding math.Abs and casting allows inlining in
func OffsetHM(d time.Duration) (offsetH, offsetM int) {
	offsetH = int(d.Hours())
	offsetM = int(d.Minutes()) % 60

	// Ensure minutes is positive
	if offsetM < 0 {
		offsetM = -offsetM
	}
	offsetM = offsetM % 60

	return
}

// LocationOffsetString get an offset in HHMM format based on hours and
// minutes offset from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
func LocationOffsetString(d time.Duration) (string, error) {
	return locationOffsetString(d, false)
}

// LocationOffsetStringDelimited get an offset in HHMM format based on hours and
// minutes offset from UTC.
//
// For 5 hours and 30 minutes
//  05:30
//
// For -5 hours and 30 minutes
//  -05:00
func LocationOffsetStringDelimited(d time.Duration) (string, error) {
	return locationOffsetString(d, true)
}

// TwoDigitOffset get digit offset for hours and minutes. This is designed
// solely to help with calculating offset strings for timestamps without using
// fmt.Sprintf, which causes allocations. This function is about 1/3 faster than
// fmt.Sprintf.
func TwoDigitOffset(in int, addPrefix bool) (digits string, err error) {
	// This is only meant to be for 2 digit offsets, such as for hours and
	// minutes offset from UTC.
	if in > 99 || in < -99 {
		err = errors.New("Out of range")
		return
	}

	// Figure out prefix based on sign of input and make input always positive
	var prefix rune = '+'
	if in < 0 {
		prefix = '-'
		in = -in
	}

	// First rune is the integer part after an integer division
	// Second rune is the remainder
	var fr rune = rune('0' + int(in/10))
	var lr rune = rune('0' + in%10)

	// Return either the prefixed two characters or non-prefixed
	if addPrefix == true {
		return utility.RunesToString(prefix, fr, lr), nil
	}
	return utility.RunesToString(fr, lr), nil
}

// OffsetString get an offset in HHMM format based on hours and minutes offset
// from UTC.
//
// For 5 hours and 30 minutes
//  0530
//
// For -5 hours and 30 minutes
//  -0500
func locationOffsetString(d time.Duration, delimited bool) (offset string, err error) {
	offsetH, offsetM := OffsetHM(d)

	xfmt := new(xfmt.Buffer)

	h, err := TwoDigitOffset(offsetH, true)
	if err != nil {
		return
	}
	xfmt.S(h)
	if delimited == true {
		xfmt.C(':')
	}
	m, err := TwoDigitOffset(offsetM, false)
	if err != nil {
		return
	}
	xfmt.S(m)

	offset = BytesToString(xfmt.Bytes()...)

	return
}

// RangeOverTimes returns a date range function over start date to end date inclusive.
// After the end of the range, the range function returns a zero date,
// date.IsZero() is true. If zones for start and end differ an error will be
// returned and needs to be checked for before time.IsZero().
//
// Note that this function has been modified to NOT change the location for the
// start and end time to UTC. This is in keeping with the avoidance of change to
// time locations passed into function. It is the responsibility of the caller
// to set location in keeping with the intended use of the function. The
// location used could affect the day values.
//
// Sample usage assuming building a map with empty string values:
/*
	t1 := time.Now()
	t2 := t1.Add(30 * 24 * time.Hour)

	m := make(map[string]string)

	var err error
	var newTime time.Time
	for rt := timestamp.RangeOverTimes(t1, t2); ; {
		newTime, err = rt()
		if err != nil {
			// Handle when there was an error in the input times
			break
		}
		if newTime.IsZero() {
			// Handle when the day range is done
			break
		}
		v := fmt.Sprintf("%04d-%02d-%02d", newTime.Year(), newTime.Month(), newTime.Day())
		m[v] = ""
	}

	if err != nil {
		// handle error due to non-equal UTC offsets
	}

	a := make([]string, 0, len(m))
	for v := range m {
		a = append(a, v)
	}
	sort.Strings(a)
	fmt.Println("Days in range")

	for _, v := range a {
		fmt.Println("Got", v)
	}
*/
func RangeOverTimes(start, end time.Time) func() (time time.Time, err error) {
	_, startZone := start.Zone()
	_, endZone := end.Zone()

	if startZone != endZone {
		return func() (time.Time, error) {
			return time.Time{}, errors.New("Zones for start and end differ")
		}
	}

	y, m, d := start.Date()
	start = time.Date(y, m, d, 0, 0, 0, 0, start.Location())
	y, m, d = end.Date()
	end = time.Date(y, m, d, 0, 0, 0, 0, end.Location())

	return func() (time.Time, error) {
		if start.After(end) {
			return time.Time{}, nil
		}
		date := start
		start = start.AddDate(0, 0, 1)

		return date, nil
	}
}

// TimeDateOnly get date with zero time values
//
// Note that this function has been modified to NOT change the location for the
// start and end time to UTC. This is in keeping with the avoidance of change to
// time locations passed into function. It is the responsibility of the caller
// to set location in keeping with the intended use of the function.
//
// Can inline
func TimeDateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// RFC7232 get format used for http headers
//   "Mon, 02 Jan 2006 15:04:05 GMT"
//
// TimeFormat is the time format to use when generating times in HTTP headers.
// It is like time.RFC1123 but hard-codes GMT as the time zone. The time being
// formatted must be in UTC for Format to generate the correct format. This is
// done in the function before the call to format.
//
// Can't inline as the format must be in GMT
func RFC7232(t time.Time) string {
	t = t.In(time.UTC)
	return t.Format(http.TimeFormat)
}

// ISO8601Compact ISO-8601 timestamp with no sub seconds
//   "20060102T150405-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601Compact(t time.Time) string {
	return t.Format("20060102T150405-0700")
}

// ISO8601CompactMsec ISO-8601 timestamp with no seconds
//   "20060102T150405.000-0700"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601CompactMsec(t time.Time) string {
	return t.Format("20060102T150405.000-0700")
}

// ISO8601 ISO-8601 timestamp long format string result
//   "2006-01-02T15:04:05-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601(t time.Time) string {
	return t.Format("2006-01-02T15:04:05-07:00")
}

// ISO8601Msec ISO-8601 longtimestamp with msec
//   "2006-01-02T15:04:05.000-07:00"
//
// Result will be in whatever the location the incoming time is set to. If UTC
// is desired set location to time.UTC first
func ISO8601Msec(t time.Time) string {
	return t.Format("2006-01-02T15:04:05.000-07:00")
}

// StartTimeIsBeforeEndTime if time 1 is before time 2 return true, else false
func StartTimeIsBeforeEndTime(t1 time.Time, t2 time.Time) bool {
	return t2.Unix()-t1.Unix() > 0
}
