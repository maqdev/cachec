package pgconvert

import "encoding/binary"

func Binary_AppendInt64(keyBytes []byte, i int64) []byte {
	return binary.LittleEndian.AppendUint64(keyBytes, uint64(i))
}
