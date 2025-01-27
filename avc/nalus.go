package avc

import (
	"encoding/binary"
	"fmt"
)

// GetNalusFromSample - get nalus by following 4 byte length fields
func GetNalusFromSample(sample []byte) ([][]byte, error) {
	length := len(sample)
	if length < 4 {
		return nil, fmt.Errorf("less than 4 bytes, No NALUs")
	}
	naluList := make([][]byte, 0, 2)
	var pos uint32 = 0
	for pos < uint32(length-4) {
		naluLength := binary.BigEndian.Uint32(sample[pos : pos+4])
		pos += 4
		if int(pos+naluLength) > len(sample) {
			return nil, fmt.Errorf("NALU length fields are bad. Not video?")
		}
		naluList = append(naluList, sample[pos:pos+naluLength])
		pos += naluLength
	}
	return naluList, nil
}

// GetSampleFromNalus - get sample from nal units
func GetSampleFromNalus(nalus [][]byte) ([]byte, error) {
	var sample []byte
	for _, nalu := range nalus {
		size := make([]byte, 4, 4)
		binary.BigEndian.PutUint32(size, uint32(len(nalu)))
		sample = append(sample, size...)
		sample = append(sample, nalu...)
	}
	return sample, nil
}