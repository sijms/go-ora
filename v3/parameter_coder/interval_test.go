package parameter_coder

import (
	"encoding/binary"
	"testing"

	"github.com/sijms/go-ora/v3/types"
)

// TestIntervalParameterUsesColumnPrecision checks that the column precision and scale set on
// the coder reach the decoded string: an INTERVAL DAY(2) TO SECOND(0) column has no fraction.
func TestIntervalParameterUsesColumnPrecision(t *testing.T) {
	raw := make([]byte, 11)
	binary.BigEndian.PutUint32(raw, uint32(0x80000000+2))
	raw[4], raw[5], raw[6] = 72, 83, 94
	binary.BigEndian.PutUint32(raw[7:], uint32(0x80000000))
	param := &IntervalParameter{BasicParameter: BasicParameter{DataType: types.INTERVALDS_DTY, BValue: raw}, Precision: 2, Scale: 0}
	got, err := param.Decode(nil)
	if err != nil || got != "+02 12:23:34" {
		t.Errorf("expected +02 12:23:34 and got %v (%v)", got, err)
	}
}
