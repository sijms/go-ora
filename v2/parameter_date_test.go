package go_ora

import (
	"bytes"
	"testing"
	"time"
	_ "time/tzdata" // the zones the tests load, wherever they run

	"github.com/sijms/go-ora/v2/converters"
	"github.com/sijms/go-ora/v2/network"
)

// dateValues is the wire form of values, each as GetClr reads it: a length byte, then
// the encoded DATE.
func dateValues(values ...time.Time) []byte {
	var input []byte
	for _, v := range values {
		data := converters.EncodeDate(v)
		input = append(input, byte(len(data)))
		input = append(input, data...)
	}
	return input
}

// decodeDates decodes len(values) values of type dataType from a connection whose
// zones are set by setZone before each value.
func decodeDates(t *testing.T, dataType TNSType, values []time.Time, setZone func(i int, conn *Connection)) []time.Time {
	t.Helper()
	conn := &Connection{session: network.NewSessionWithInputBufferForDebug(dateValues(values...))}
	out := make([]time.Time, len(values))
	for i := range values {
		setZone(i, conn)
		par := ParameterInfo{DataType: dataType}
		if err := par.decodePrimValue(conn, nil, false); err != nil {
			t.Fatalf("value %d: %v", i, err)
		}
		got, ok := par.oPrimValue.(time.Time)
		if !ok {
			t.Fatalf("value %d: %T, want time.Time", i, par.oPrimValue)
		}
		out[i] = got
	}
	return out
}

// A DATE or TIMESTAMP keeps its wall clock and takes the server's zone, unless that
// zone is UTC by name and offset: then it stays in time.UTC. A TIMESTAMP WITH LOCAL TIME
// ZONE does the same with the database's zone. Zones named "UTC" that are not time.UTC
// are the case of #686 and #694; the zone also changes between values.
func TestDateValuesTakeTheConnectionsZone(t *testing.T) {
	wall := time.Date(2024, 6, 1, 12, 30, 15, 0, time.UTC)
	zones := []struct {
		loc *time.Location
		utc bool // the value stays in time.UTC
	}{
		{time.UTC, true},
		{time.FixedZone("", 4*60*60), false},
		{time.FixedZone("UTC", 0), true},
		{time.FixedZone("", 0), false},
		{time.FixedZone("UTC", 60*60), false},
		{time.UTC, true},
	}
	etc, err := time.LoadLocation("Etc/UTC")
	if err != nil {
		t.Fatal(err)
	}
	zones = append(zones, struct {
		loc *time.Location
		utc bool
	}{etc, true})
	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	zones = append(zones, struct {
		loc *time.Location
		utc bool
	}{ny, false})
	testScenarios := []struct {
		description string
		dataType    TNSType
		setZone     func(conn *Connection, loc *time.Location)
	}{
		{description: "DATE", dataType: DATE, setZone: func(conn *Connection, loc *time.Location) { conn.dbServerTimeZone = loc }},
		{description: "TIMESTAMP", dataType: TIMESTAMP, setZone: func(conn *Connection, loc *time.Location) { conn.dbServerTimeZone = loc }},
		{description: "TIMESTAMP WITH LOCAL TIME ZONE", dataType: TimeStampeLTZ, setZone: func(conn *Connection, loc *time.Location) { conn.dbTimeZone = loc }},
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.description, func(t *testing.T) {
			values := make([]time.Time, len(zones))
			for i := range values {
				values[i] = wall
			}
			got := decodeDates(t, scenario.dataType, values, func(i int, conn *Connection) { scenario.setZone(conn, zones[i].loc) })
			for i, g := range got {
				wantLoc := zones[i].loc
				if zones[i].utc {
					wantLoc = time.UTC
				}
				want := time.Date(2024, 6, 1, 12, 30, 15, 0, wantLoc)
				if !g.Equal(want) || g.Location() != wantLoc {
					t.Errorf("value %d, zone %q: %v (time.UTC: %v), want %v (time.UTC: %v)", i, zones[i].loc, g, g.Location() == time.UTC, want, wantLoc == time.UTC)
				}
			}
		})
	}
}

// BenchmarkDecodeDate decodes DATE values from a server in UTC, and in another zone.
func BenchmarkDecodeDate(b *testing.B) {
	for _, zone := range []struct {
		name string
		loc  *time.Location
	}{{"UTC", time.UTC}, {"+04:00", time.FixedZone("", 4*60*60)}} {
		b.Run(zone.name, func(b *testing.B) {
			input := bytes.Repeat(dateValues(time.Date(2024, 6, 1, 12, 0, 15, 0, time.UTC)), b.N)
			conn := &Connection{session: network.NewSessionWithInputBufferForDebug(input), dbServerTimeZone: zone.loc}
			par := ParameterInfo{DataType: DATE}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := par.decodePrimValue(conn, nil, false); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
