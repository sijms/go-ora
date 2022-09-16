// 2022/9/16 Bin Liu <bin.liu@enmotech.com>

package converters

import (
	"testing"
)

func Benchmark_StringConverter_Decode(b *testing.B) {
	conv := NewStringConverter(853)
	data := []byte("dogThe quick brown fox jumps over the lazy")
	for i := 0; i < b.N; i++ {
		_ = conv.Decode(data)
	}
}
