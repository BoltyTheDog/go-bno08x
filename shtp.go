package bno08x

import (
	"encoding/binary"
	"fmt"
)

const HeaderLen = 4

// SHTPHeader represents the 4-byte SHTP header
type SHTPHeader struct {
	Length         uint16
	Channel        uint8
	SequenceNumber uint8
}

// ParseHeader parses a 4-byte buffer into an SHTPHeader
func ParseHeader(data []byte) (SHTPHeader, error) {
	if len(data) < HeaderLen {
		return SHTPHeader{}, fmt.Errorf("buffer too short for header")
	}
	length := binary.LittleEndian.Uint16(data[0:2])
	// The MSB of length is used for continuity in some cases, but BNO08x uses it differently
	// According to Hillcrest, we should mask out the top bit
	length &= 0x7FFF

	return SHTPHeader{
		Length:         length,
		Channel:        data[2],
		SequenceNumber: data[3],
	}, nil
}

// EncodeHeader encodes the header into the provided 4-byte buffer
func (h SHTPHeader) Encode(data []byte) {
	binary.LittleEndian.PutUint16(data[0:2], h.Length)
	data[2] = h.Channel
	data[3] = h.SequenceNumber
}
