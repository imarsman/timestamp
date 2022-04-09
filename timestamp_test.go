package timestamp_test

import (
	"fmt"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/imarsman/timestamp"
	"github.com/imarsman/timestamp/pkg/xfmt"
	"github.com/matryer/is"
)

//                Tests and benchmarks
// -----------------------------------------------------
// benchmark
//   go test -run=XXX -bench=. -benchmem
// Get allocation information and pipe to less
//   go build -gcflags '-m -m' ./*.go 2>&1 |less
// Run all tests
//   go test -v
// Run one test and do allocation profiling
//   go test -run=XXX -bench=IterativeISOTimestampLong -gcflags '-m' 2>&1 |less
// Run a specific test by function name pattern
//  go test -run=TestParsISOTimestamp
//
//  go test -run=XXX -bench=.
//  go test -bench=. -benchmem -memprofile memprofile.out -cpuprofile cpuprofile.out
//  go tool pprof -http=:8080 memprofile.out
//  go tool pprof -http=:8080 cpuprofile.out

//report start of running time tracking
func runningtime(s string) (string, time.Time) {
	fmt.Println("Start:	", s)
	return s, time.Now()
}

// report total running time
func track(s string, startTime time.Time) {
	endTime := time.Now()
	fmt.Println("End:	", s, "took", endTime.Sub(startTime))
}

// sample running time tracking
func execute() {
	defer track(runningtime("execute"))
	time.Sleep(3 * time.Second)
}

// For use in parse checking. If the timestamp has no zone use the location
// passed in. Compare expected to the parssed value in the location pased in.
func checkDate(t *testing.T, input string, location *time.Location) (
	got string, parsed string, calculated time.Time, calculatedOffset time.Duration, defaultOffset time.Duration) {
	is := is.New(t)

	calculated, err := timestamp.ParseInLocation(input, location)
	if err != nil {
		t.Logf("Got error on parsing %v", err)
	}
	is.NoErr(err)

	parsed = timestamp.ISO8601Msec(calculated)

	calculatedOffset = timestamp.OffsetForTime(calculated)

	inLoc := calculated.In(location)
	defaultOffset, err = timestamp.OffsetForLocation(
		inLoc.Year(), inLoc.Month(), inLoc.Day(), inLoc.Location().String())
	is.NoErr(err)

	fmt.Printf("Input %s, calculated %v calculatedOffset %v defaultOffset %v\n", input, parsed, calculatedOffset, defaultOffset)

	return input, parsed, calculated, calculatedOffset, defaultOffset
}

// TestParse parse all patterns and compare with expected values. Input
// tmestamps are parsed and have the timestamp value applied if available,
// otherwise the passed in location is used.
func TestParse(t *testing.T) {
	is := is.New(t)

	mst, err := time.LoadLocation("MST")
	is.NoErr(err) // Location should parse without error
	jst, err := time.LoadLocation("Asia/Tokyo")
	is.NoErr(err) // Location should parse without error

	test, err := time.LoadLocation("EST")
	is.NoErr(err) // Location should parse without error
	est, err := time.LoadLocation("America/Toronto")
	is.NoErr(err)        // Location should parse without error
	is.True(test != est) // EST should not equal America/Toronto

	utcST, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error
	utcST = utcST.In(time.UTC)
	utcOffset, err := timestamp.OffsetForLocation(utcST.Year(), utcST.Month(), utcST.Day(), time.UTC.String())
	is.NoErr(err) // Getting offset should not have resulted in error

	estST, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error
	estST = estST.In(est)
	estSTOffset, err := timestamp.OffsetForLocation(estST.Year(), estST.Month(), estST.Day(), est.String())
	is.NoErr(err) // Getting offset should not have resulted in error

	estDST, err := timestamp.ParseInUTC("2006-07-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error
	estDST = estDST.In(est)
	estDSTOffset, err := timestamp.OffsetForLocation(estDST.Year(), estDST.Month(), estDST.Day(), est.String())
	is.NoErr(err) // Getting offset should not have resulted in error

	mstST, err := timestamp.ParseInUTC("2006-07-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error
	mstST = mstST.In(mst)
	mstOffset, err := timestamp.OffsetForLocation(mstST.Year(), mstST.Month(), mstST.Day(), mst.String())
	is.NoErr(err) // Getting offset should not have resulted in error

	// Tokyo time test
	jST, err := timestamp.ParseInUTC("2006-07-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error
	jST = mstST.In(mst)
	jstOffset, err := timestamp.OffsetForLocation(jST.Year(), jST.Month(), jST.Day(), jst.String())
	is.NoErr(err) // Getting offset should not have resulted in error

	var start = time.Now() // Start the timing of time to do tests

	// It is possible to have a string which is just digits that will be parsed
	// as a timestamp, incorrectly.
	_, err = timestamp.ParseInUTC("2006010247")
	is.NoErr(err) // Timestamp with just date should parse without error

	// Get a unix timestamp we should not parse
	_, err = timestamp.ParseInUTC("1")
	is.True(err != nil) // Error should be true

	// Get time value from parsed reference time
	unixBase, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse without error

	var sent, tStr string
	var res time.Time
	var resOffset, defOffset, d time.Duration

	is.True(sent == "")
	is.True(tStr == "")
	is.True(res == time.Time{})

	// This could be parsed as ISO-8601
	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.UnixNano()), time.UTC)
	is.Equal(resOffset, defOffset)
	// UTC timestamp stayed UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.Unix()), time.UTC)
	is.Equal(resOffset, defOffset)
	// UTC timestamp converted to MST
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, fmt.Sprint(unixBase.Unix()), mst)
	is.Equal(resOffset, defOffset)
	is.Equal(mstOffset, resOffset)

	// Handling of a leap second. The minute value rolls over to the next minute
	// when the second value is 60.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150460-0700", time.UTC)
	is.True(resOffset != defOffset)

	// Should be offset corresponding to Mountain Standard Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150460-0700", mst)
	is.Equal(resOffset, defOffset)
	is.Equal(mstOffset, resOffset)

	// Should be offset corresponding to Eastern Standard time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102240000-0500", est)
	is.Equal(resOffset, defOffset)
	// Result offset is EST
	is.Equal(estSTOffset, resOffset)

	// Should be offset corresponding to Tokyo time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102240000+0900", jst)
	is.Equal(resOffset, defOffset)
	// Result offset is JST
	is.Equal(jstOffset, resOffset)

	// Should be offset corresponding to Tokyo time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102240000+0900", est)
	is.True(resOffset != defOffset)
	// Result offset is JST
	is.Equal(jstOffset, resOffset)

	// Should be offset corresponding to Estern Daylight Savings Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060702240000-0400", est)
	// Is DST
	is.Equal(estDSTOffset, resOffset)

	// Short ISO-8601 timestamps with numerical zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000+0000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000-0000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000+0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is not MST
	is.True(mstOffset != resOffset)
	// Result is UTC+07:00
	d, err = time.ParseDuration("7h")
	is.NoErr(err)
	is.True(resOffset == d)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000+0330", time.UTC)
	is.True(resOffset != defOffset)
	// Result is UTC+03:30
	d, err = time.ParseDuration("3h30m")
	is.NoErr(err)
	is.True(resOffset == d)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Long ISO-8601 timestamps with numerical zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.001000-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.001000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.999999999-07", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Short  ISO-8601 timestamps with UTC zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.001000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000100000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Long date time with UTC zone offsets
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.000000Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15:04:05.999999999Z", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Just in case
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Short ISO-8601 timestamps with no zone offset. Assume UTC.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.000000", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102T150405.999999999", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// SQL
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", time.UTC)
	is.Equal(utcOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// MST is -0700 from UTC, so UTC will be 7 hours ahead
	// Should be offset corresponding to Mountain Standard Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", mst)
	is.Equal(resOffset, defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// MST is -0500 from UTC, so UTC will be 5 hours ahead
	// Should be offset corresponding to Eastern Daylight Savings Time
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05", est)
	is.Equal(resOffset, defOffset)
	// Result is EST
	is.Equal(estSTOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// The input has a timestamp so EST will not be applied
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00", est)
	is.True(resOffset != defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 +00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 -00:00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 22:04:05 +00:00", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Hopefully less likely to be found. Assume UTC.
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006/01/02", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "01/02/2006", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "1/2/2006", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is UTC
	is.Equal(utcOffset, resOffset)

	// Weird ones with improper separators
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000000-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.999999999-0700", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.000000-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02T15-04-05.999999999-07:00", time.UTC)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(resOffset, mstOffset)

	// RFC7232 - used in HTTP protocol
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 GMT", time.UTC)
	is.Equal(resOffset, defOffset)
	// Result is MST
	is.Equal(utcOffset, resOffset)

	// RFC1123Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05", est)
	is.Equal(estSTOffset, resOffset)
	// Result is EST
	is.Equal(resOffset, defOffset)

	// RFC822Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "02 Jan 06 15:04 -0700", time.UTC)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.True(resOffset != defOffset)

	// Just in case
	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Try modifying zone
	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Will be offset 7 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", mst)
	is.Equal(mstOffset, resOffset)
	// Result is MST
	is.Equal(resOffset, defOffset)

	// Try modifying zone
	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0700", est)
	is.True(resOffset != defOffset)
	// Result is MST
	is.Equal(mstOffset, resOffset)

	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", est)
	is.Equal(estSTOffset, resOffset)
	// Result is EST
	is.Equal(resOffset, defOffset)

	// EST not used because a different offset is in the timestamp
	sent, tStr, res, resOffset, defOffset = checkDate(t, "Mon, 02 Jan 2006 15:04:05 -0600", mst)
	is.True(resOffset != defOffset)
	// Result is UTC-06:00
	d, err = time.ParseDuration("-6h")
	is.NoErr(err)
	is.True(resOffset == d)

	// RFC822Z
	sent, tStr, res, resOffset, defOffset = checkDate(t, "02 Jan 06 15:04 -0700", est)
	is.True(resOffset != defOffset)
	is.Equal(mstOffset, resOffset)

	// Just in case
	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "2006-01-02 15-04-05", est)
	is.Equal(estSTOffset, resOffset)
	is.Equal(resOffset, defOffset)

	// Will be offset 5 hours to get UTC
	sent, tStr, res, resOffset, defOffset = checkDate(t, "20060102150405", est)
	is.Equal(estSTOffset, resOffset)
	is.Equal(resOffset, defOffset)

	t.Logf("Took %v to check", time.Since(start))
}

func TestISOCompare(t *testing.T) {
	is := is.New(t)

	start := time.Now()
	// It is possible to have a strring which is just digits that will be parsed
	// as a timestamp, incorrectly.

	ts := "2006-01-02T15:04:05-07:00"
	_, err := timestamp.ParseISOInUTC(ts)
	is.NoErr(err)
	count := 1000

	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseInUTC(ts)
		is.NoErr(err) // Timestamp should parse with no error
	}

	t.Logf("Took %v to parse %s %d times", time.Since(start), ts, count)

	start = time.Now()

	ts = "20060102T150405-0700"

	for i := 0; i < count; i++ {
		// Get a unix timestamp we should not parse
		_, err := timestamp.ParseInUTC(ts)
		is.NoErr(err) // Timestamp should parse with no error
	}

	t.Logf("Took %v to parse %s %d times", time.Since(start), ts, count)
}

// Note that the range of days returned by RangeOverTimes will result in a span
// from the start time to the end time, which will be one more than the number
// of days added to the start time.
func TestRangeOverTimes(t *testing.T) {
	is := is.New(t)

	t1 := time.Now()
	// Make zone incompatible to test error
	t2 := t1.Add(10 * 24 * time.Hour).In(time.UTC)

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
	is.True(err != nil)
	fmt.Println("Got error", err)

	t2 = t1.Add(10 * 24 * time.Hour)

	for rd := timestamp.RangeOverTimes(t1, t2); ; {
		newTime, err = rd()
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
	is.NoErr(err) // There should not have been an error in getting range of dates
	t1 = time.Now()
	t2 = t1.Add(10 * 24 * time.Hour)

	a := make([]string, 0, len(m))
	for v := range m {
		a = append(a, v)
	}
	sort.Strings(a)

	fmt.Println("Days in range")
	for _, v := range a {
		fmt.Println("Got", v)
	}
}

// TestOrdering check ordering call
func TestOrdering(t *testing.T) {
	is := is.New(t)

	t1, err1 := timestamp.ParseInUTC("20201210T223900-0500")
	is.NoErr(err1) // Timestamp should parse with no error

	t2, err2 := timestamp.ParseInUTC("20201211T223900-0500")
	is.NoErr(err2) // Timestamp should parse with no error

	is.True(timestamp.StartTimeIsBeforeEndTime(t1, t2))  // Start is before end
	is.True(!timestamp.StartTimeIsBeforeEndTime(t2, t1)) // Start is not before end
}

// Test how long it take to parse a timestamp 1,000 times
func TestTime(t *testing.T) {
	is := is.New(t)

	var unixBase time.Time
	var err error
	count := 1000

	defer track(runningtime(fmt.Sprintf("Time to parse timestamp %dx", count)))

	for i := 0; i < count; i++ {
		unixBase, err = timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	}

	is.NoErr(err) // Timestamp should parse with no error
	t.Logf("Timestamp %s", timestamp.ISO8601Msec(unixBase))
}

func TestFormat(t *testing.T) {
	is := is.New(t)
	ts, err := timestamp.ParseInUTC("2006-01-02T15:04:05.000+00:00")
	is.NoErr(err) // Timestamp should parse with no error

	var s string
	count := 1000

	defer track(runningtime(fmt.Sprintf("Time to format timestamp %dx", count)))

	for i := 0; i < count; i++ {
		s = timestamp.ISO8601Msec(ts)
	}

	t.Logf("Timestamp %s", s)
}

var locations = []string{
	"MST",
	"America/New_York",
	"UTC",
	"Asia/Kabul",
	"America/St_Johns",
	"Europe/London",
	"America/Argentina/San_Luis",
	"Canada/Newfoundland",
	"Asia/Calcutta",
	"Asia/Tokyo",
	"America/Toronto",
}

// TestOffsetForZones test to get the offset for dates with a named zone. This
// could be more accurately done by removing the zone information from the
// timestamp string but normally this sort of opperation would be needed when
// timetamps were available without zone information but the location was known.
func TestOffsetForZones(t *testing.T) {
	is := is.New(t)

	// var hours, minutes int
	var err error
	t1, err := timestamp.ParseInUTC("20200101T000000Z")
	is.NoErr(err) // Timestamp should parse without error
	t2, err := timestamp.ParseInUTC("20200701T000000Z")
	is.NoErr(err) // Timestamp should parse without error
	defer track(runningtime(fmt.Sprintf("Time to get offset information for %d locations/dates", len(locations)*2)))
	for _, location := range locations {
		for _, tNext := range []time.Time{t1, t2} {
			d, err := timestamp.OffsetForLocation(tNext.Year(), tNext.Month(), tNext.Day(), location)
			is.NoErr(err) // Should be no error getting offset for location
			offset, err := timestamp.LocationOffsetStringDelimited(d)
			is.NoErr(err)
			fmt.Printf("zone %s time %v offset %s\n", location, tNext, offset)
		}
	}
}

// Test how long it takes to get timezone information 1,000 times.
func TestZoneTime(t *testing.T) {
	is := is.New(t)

	zone := "Canada/Newfoundland"
	count := 1000

	var d time.Duration
	var offset string
	var err error

	offset, err = timestamp.LocationOffsetString(d)
	t.Logf("Offset %s", offset)
	offset, err = timestamp.LocationOffsetStringDelimited(d)
	t.Logf("Offset delimited %s", offset)
	is.NoErr(err)

	defer track(runningtime(fmt.Sprintf("Time to get zone information %dx with zero allocation", count)))

	d, err = timestamp.OffsetForLocation(2006, 1, 1, zone)
	for i := 0; i < count; i++ {
		is.NoErr(err) // There should not have been an error
		offset, err = timestamp.LocationOffsetString(d)
		is.NoErr(err)
	}

	t.Logf("start zone %s offset %s hours %d minutes %d offset %s error %v",
		zone, offset, int(d.Hours()), int(d.Minutes()), offset, err)
}

func TestZoneTimeFmt(t *testing.T) {
	is := is.New(t)

	zone := "Canada/Newfoundland"
	count := 1000

	var d time.Duration
	var offset string
	var err error

	d, err = timestamp.OffsetForLocation(2006, 1, 1, zone)
	offsetH, offsetM := timestamp.OffsetHM(d)

	defer track(runningtime(fmt.Sprintf("Time to get zone information %dx with allocation", count)))

	for i := 0; i < count; i++ {
		is.NoErr(err) // There should not have been an error
		offset = fmt.Sprintf("%02d%02d", offsetH, offsetM)
		is.NoErr(err)
	}

	t.Logf("Offset %s", offset)

	t.Logf("start zone %s offset %s hours %d minutes %d offset %s error %v",
		zone, offset, int(d.Hours()), int(d.Minutes()), offset, err)
}

// Test conversion of string to int
// func TestStringToInt(t *testing.T) {
// 	is := is.New(t)
// 	result, err := timestamp.StringToInt("123456789")
// 	is.NoErr(err)
// 	t.Log(result)
// 	result, err = timestamp.StringToInt("1234567090")
// 	is.True(err != nil)
// 	t.Log("Got error - expected", err)
// }

// Benchmark the string to int function
// func BenchmarkStringToInt(b *testing.B) {
// 	is := is.New(b)

// 	var err error
// 	var result int

// 	b.SetBytes(bechmarkBytesPerOp)
// 	b.ReportAllocs()
// 	b.SetParallelism(30)
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			result, err = timestamp.StringToInt("123456789")
// 		}
// 	})

// 	is.True(result != 0) // Should not be default null/0
// 	is.NoErr(err)        // Parsing should not have caused an error
// }

// Benchmark strconv.Atoi for comparison with the custom function
func BenchmarkStringToIntAtoi(b *testing.B) {
	is := is.New(b)

	var err error
	var result int

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			result, err = strconv.Atoi("123456789")
		}
	})

	is.True(result != 0) // Should not be default null/0
	is.NoErr(err)        // Parsing should not have caused an error
}

// Test seprate call to parse a Unix timestamp.
func TestParseUnixTimestamp(t *testing.T) {
	is := is.New(t)

	var err error
	var t1, t2 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.UnixNano())
	t.Logf("Nano timestamp string %s len %d", ts1, len(ts1))
	ts2 := fmt.Sprint(now.Unix())

	count := 1000

	defer track(runningtime(fmt.Sprintf("Time to parse two timestamps %dx", count*2)))

	for i := 0; i < count; i++ {
		t1, err = timestamp.ParseUnixTS(ts1)
		is.NoErr(err) // Should have been no error
		t2, err = timestamp.ParseUnixTS(ts2)
		is.NoErr(err) // Should have been no error
	}

	is.True(t1 != time.Time{}) // Should not be empty time
	is.True(t2 != time.Time{}) // Should not be empty time
}

func TestParseLocation(t *testing.T) {
	t1, _ := time.Parse("2006-01-02T15:04:05-0700", "2006-02-02T15:04:05-0700")

	t.Logf("Got time %v at location %s", t1.Format(time.RFC1123), t1.Location())
}

func TestGetDigits(t *testing.T) {
	is := is.New(t)

	var s string
	var err error
	for i := 0; i < 1000; i++ {
		s, err = timestamp.TwoDigitOffset(1, true)
		is.NoErr(err)
		s, err = timestamp.TwoDigitOffset(-1, true)
		is.NoErr(err)
	}
	t.Log(s)

}
func TestParsISOTimestamp(t *testing.T) {
	is := is.New(t)

	var err error
	count := 1000
	var ts time.Time

	formats := []string{
		"20060102T010101",
		"2006-13-02T40:01:01.123456789+01:00",
		"2006-13-02T40:01:01.999999999+01:00",
		"20060102T010101.123456789",
		"20060102T010101.12345678",
		"20060102T010101.1234567",
		"20060102T010101.123456",
		"20060102T010101.123456-0500",
		"20060102T010101-0400",
		"20060102t010101-0400",
		"20060102T010101+0400",
		// Force a rollover of day
		// This is part of the functionality of the Go time library
		"20060102T240000+0400",
		"2006-01-02T01:01:01-04:00",
		"2006-01-02T01:01:01-03:30",
		"2006-01-02T01:01:01-04:30",
		"2006-01-02T18:01:01+01:00",
		"2006-01-02T18-01-01+0100",
		// Hour will force rollover of day
		// This is part of the functionality of the Go time library
		"2006-01-02T27-01-01+0100",
		// Oddball separators but still can parse
		"2006/01/02T18.01.01+01:00",
		// Day will force rollover of month
		// This is part of the functionality of the Go time library
		"2000-02-30T12-01-01+0100",
	}

	badFormats := []string{
		// Bad minutes offset
		"2006/01/02T18.01.01+01:10",
		// Invalid offset digit count
		"2006-01-02T11-30-61+010",
		"bkjfdlkjdfsaj;g;lkjafdkljl;fdaladf;jkladfsl;kfjads;j",
		// Time zone too many characters
		"20060102T010101-04000",
		// Two zulu indicators
		"20060102T0101Z01Z",
		// More bad characters
		"2006w01s02T18a01b01c01:00",
		// More bad characters
		"2006-01-02T18:01:01b01:00",
	}

	t.Log("Correct intput")
	for _, in := range formats {
		ts, err = timestamp.ParseISOTimestamp(in, time.UTC)
		t.Logf("input %s ts %v", in, ts)
		is.NoErr(err) // Should be no error
	}

	t.Log("")
	t.Log("Invalid intput")
	for _, in := range badFormats {
		ts, err = timestamp.ParseISOTimestamp(in, time.UTC)
		t.Logf("input %s error %v", in, err)
		is.True(err != nil) // Should be an error
	}

	defer track(runningtime(fmt.Sprintf("Time to process ISO timestamp %dx", count*2)))
	for i := 0; i < count; i++ {
		ts, err = timestamp.ParseISOTimestamp("20060102T010101.", time.UTC)
		is.NoErr(err) // Should be no error
	}
	t.Log("ts", ts)
}

const bechmarkBytesPerOp int64 = 10

func BenchmarkTwoDigitOffsets(b *testing.B) {
	is := is.New(b)

	var err error
	var s string

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s, err = timestamp.TwoDigitOffset(-7, true)
		}
	})

	is.True(s != "")
	is.NoErr(err)
}

func BenchmarkTwoDigitOffsetsFmt(b *testing.B) {
	// is := is.New(b)

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = fmt.Sprintf("%03d", -1)
		}
	})

}

// Benchmark parsing a unix timestamp
func BenchmarkUnixTimestamp(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.Unix())

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseUnixTS(ts1)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark a unix nano timestamp
func BenchmarkUnixTimestampNano(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	now := time.Now()
	ts1 := fmt.Sprint(now.UnixNano())

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseUnixTS(ts1)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark a timestamp that is only a date
func BenchmarkIterativeISOTimestampDateOnly(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseISOTimestamp("2006-07-02", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark a timestamp with no delimiters or zone
func BenchmarkIterativeISOTimestampShortNoZone(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseISOTimestamp("20060102T010101", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark a timestamp with no delimiters but a zone
func BenchmarkIterativeISOTimestampCompactNoMsecWithZone(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseISOTimestamp("20060702T010101+0130", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// The most computational and allocationally intensive timestamp to parse, with
// a nonzero value in every part plus delimiters.
func BenchmarkIterativeISOTimestampMsecAllPartsNonzero(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Max allocations with no part zero
			t1, err = timestamp.ParseISOTimestamp("2006-07-02T07:01:01.999+03:30", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	// b.Log(t1)
	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// The most computational and allocationally intensive timestamp to parse, with
// a nonzero value in every part plus delimiters.
func BenchmarkIterativeISOTimestampLongAllPartsNonzero(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Max allocations with no part zero
			t1, err = timestamp.ParseISOTimestamp("2006-07-02T07:01:01.999999999+03:30", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	// b.Log(t1)
	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark the Go time parsing call with format
func BenchmarkIterativeNativeEquivalent(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = timestamp.ParseISOTimestamp("2006-01-02T15:04:05-07:00", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark the Go time parsing call with format
func BenchmarkNativeISOTimestampLong(b *testing.B) {
	is := is.New(b)

	var err error
	var t1 time.Time

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			t1, err = time.ParseInLocation("2006-01-02T15:04:05-07:00", "2006-07-02T01:01:01+01:30", time.UTC)
			if err != nil {
				b.Log(err)
			}
		}
	})

	is.True(t1 != time.Time{}) // Should not have an empty time
	is.NoErr(err)              // Parsing should not have caused an error
}

// Benchmark a non allocating buffer that is an alternative to the fmt package
func BenchmarkNonAllocatingBuffer(b *testing.B) {
	is := is.New(b)

	var s []byte
	first := "first"
	second := "second"
	third := "second"

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var xfmtBuf = new(xfmt.Buffer)
			// Avoid heap allocation for each append which uses a variable
			// The total in tests took about 80 ns/op, which is about 20 ms per
			// append. The benchmark reported 0 B/op.
			xfmtBuf.S("Could not parse as ISO timestamp ").S(first).C(' ').S(second).C(' ').S(third)
			s = xfmtBuf.Bytes()
			xfmtBuf.Reset()
		}
	})

	is.True(len(s) > 0)
}

// Benchmlark allocating fmt call
func BenchmarkAllocatingBuffer(b *testing.B) {
	is := is.New(b)

	var s string
	first := "first"
	second := "second"
	third := "second"

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Each append uses an allocation
			// The total in tests took about 80 ns/op, which is about 20 ns per
			// item. The benchmark reported 112 B/op.
			s = fmt.Sprintf("Could not parse as ISO timestamp %s %s %s", first, second, third)
		}
	})

	is.True(s != "")
}

// The goal of using strings.Builder is to avoid heap allocation
// The memory used and time taken should be similar to using a string cast
func BenchmarkBytesToString(b *testing.B) {
	is := is.New(b)

	var s string
	bytes := []byte{'a', 'b', 'c', 'd'}

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s = timestamp.BytesToString(bytes...)
		}
	})

	is.True(s != "")
}

// Benchmark creating a string from bytes using Go cast
func BenchmarkBytesToStringCast(b *testing.B) {
	is := is.New(b)

	var s string
	bytes := []byte{'a', 'b', 'c', 'd'}

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s = string(bytes)
		}
	})

	is.True(s != "")
}

// The goal of using strings.Builder is to avoid heap allocation
// The memory used and time taken should be similar to using a string cast
// func BenchmarkRunesToString(b *testing.B) {
// 	is := is.New(b)

// 	var s string
// 	runes := []rune{'a', 'b', 'c', 'd'}

// 	b.SetBytes(bechmarkBytesPerOp)
// 	b.ReportAllocs()
// 	b.SetParallelism(30)
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			s = timestamp.RunesToString(runes...)
// 		}
// 	})

// 	is.True(s != "")
// }

// Benchmark creating a string from runes using Go cast
func BenchmarkRunesToStringCast(b *testing.B) {
	is := is.New(b)

	var s string
	runes := []rune{'a', 'b', 'c', 'd'}

	b.SetBytes(bechmarkBytesPerOp)
	b.ReportAllocs()
	b.SetParallelism(30)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s = string(runes)
		}
	})

	is.True(s != "")
}
