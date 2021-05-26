package security

type DiffieHellman struct {
	buffer_1 []byte
	buffer_2 []byte
	size_1   int
	size_2   int
}

func NewDiffieHellman(buffer1, buffer2 []byte, size1, size2 int) *DiffieHellman {
	return &DiffieHellman{
		buffer_1: buffer1,
		buffer_2: buffer2,
		size_1:   size1,
		size_2:   size2,
	}
}

//func (dh *DiffieHellman) GetPublicKey() {
//	array_1 := make([]int, 257)
//	array_2 := make([]int, 257)
//	array_3 := make([]uint8, 512)
//	num := (dh.size_2 + 7) >> 3
//	m := (dh.size_1 + 7) >> 3
//	n := (dh.size_1 / 16) + 1
//	l := make([]byte, m)
//
//	// calling function b(array_3, num)
//}
