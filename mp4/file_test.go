package mp4

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/edgeware/mp4ff/bits"
)

func TestDecodeFileWithLazyMdatOption(t *testing.T) {

	// load a segment
	file, err := os.Open("./testdata/1.m4s")
	if err != nil {
		t.Error(err)
	}

	parsedFile, err := DecodeFile(file, WithDecodeMode(DecModeLazyMdat))
	if err != nil {
		t.Error(err)
	}

	for _, seg := range parsedFile.Segments {
		for _, frag := range seg.Fragments {
			if frag.Mdat.GetLazyDataSize() == 0 {
				t.Error("lazyDataSize is expected to be greater than 0")
			}
			if frag.Mdat.Data != nil {
				t.Error("Mdat Data is expected to be nil")
			}
		}
	}

}

func TestDecodeFileWithNoLazyMdatOption(t *testing.T) {

	// load a segment
	file, err := os.Open("./testdata/1.m4s")
	if err != nil {
		t.Error(err)
	}

	parsedFile, err := DecodeFile(file)
	if err != nil {
		t.Error(err)
	}

	for _, seg := range parsedFile.Segments {
		for _, frag := range seg.Fragments {
			if frag.Mdat.lazyDataSize != 0 {
				t.Error("decLazyDataSize is expected to be 0")
			}
			if frag.Mdat.Data == nil || len(frag.Mdat.Data) == 0 {
				t.Error("Mdat Data is expected to be non-nil")
			}
		}
	}
}

// TestCopyTrackSampleData checks that full early read and lazy with and without workSpace gives good and same result.
func TestCopyTrackSampleData(t *testing.T) {
	// load a progressive file
	testCases := []struct {
		lazy          bool
		workSpaceSize int
	}{
		{lazy: false, workSpaceSize: 0},
		{lazy: true, workSpaceSize: 0},
		{lazy: true, workSpaceSize: 256},
	}
	sampleDataRead := make([][]byte, 0, len(testCases))
	for j, tc := range testCases {
		fd, err := os.Open("./testdata/prog_8s.mp4")
		if err != nil {
			t.Error(err)
		}
		defer fd.Close()
		var mp4f *File
		var workSpace []byte
		if tc.lazy {
			mp4f, err = DecodeFile(fd, WithDecodeMode(DecModeLazyMdat))
			workSpace = make([]byte, tc.workSpaceSize)
		} else {
			mp4f, err = DecodeFile(fd)
		}
		if err != nil {
			t.Error(err)
		}
		var startSampleNr uint32 = 31
		var endSampleNr uint32 = 60

		for _, trak := range mp4f.Moov.Traks {
			totSize := 0
			stsz := trak.Mdia.Minf.Stbl.Stsz
			for i := startSampleNr; i <= endSampleNr; i++ {
				totSize += int(stsz.GetSampleSize(int(i)))
			}
			sampleData := bytes.Buffer{}

			err := mp4f.CopySampleData(&sampleData, fd, trak, startSampleNr, endSampleNr, workSpace)
			if err != nil {
				t.Error(err)
			}
			if sampleData.Len() != int(totSize) {
				t.Errorf("Got %d bytes instead of %d", sampleData.Len(), totSize)
			}
			if trak.Tkhd.TrackID == 1 {
				sampleDataRead = append(sampleDataRead, sampleData.Bytes())
				if len(sampleDataRead) > 1 {
					if res := bytes.Compare(sampleDataRead[j], sampleDataRead[0]); res != 0 {
						t.Errorf("sample data read differs %d", res)
					}
				}
			}
		}
	}
}

func TestDecodeEncodeProgressiveSliceWriter(t *testing.T) {
	// load a segment
	rawInput, err := ioutil.ReadFile("./testdata/prog_8s.mp4")
	if err != nil {
		t.Error(err)
	}
	rawOutput := make([]byte, len(rawInput))
	inBuf := bytes.NewBuffer(rawInput)
	parsedFile, err := DecodeFile(inBuf)
	if err != nil {
		t.Error(err)
	}
	sw := bits.NewFixedSliceWriterFromSlice(rawOutput)
	err = parsedFile.EncodeSW(sw)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(rawOutput, rawInput) {
		t.Errorf("output differs from input")
	}
}

func TestDecodeEncodeMultiSidxSegment(t *testing.T) {
	rawInput, err := ioutil.ReadFile("./testdata/multi_sidx_segment.m4s")
	if err != nil {
		t.Error(err)
	}
	rawOutput := make([]byte, len(rawInput))
	inBuf := bytes.NewBuffer(rawInput)
	parsedFile, err := DecodeFile(inBuf)
	if err != nil {
		t.Error(err)
	}
	sw := bits.NewFixedSliceWriterFromSlice(rawOutput)
	err = parsedFile.EncodeSW(sw)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Equal(rawOutput, rawInput) {
		t.Errorf("output differs from input")
	}
}
