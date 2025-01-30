package tar

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// Align adds a dummy PAXRecord field to theh header to ensure the
// next file contents start aligned with the Blksize for the `dest`
// file.
func Align(dest *os.File, hdr *Header) error {
	fi, err := dest.Stat()
	if err != nil {
		return err
	}
	st := fi.Sys().(*syscall.Stat_t)
	return alignTo(hdr, st.Size, st.Blksize)
}

const paxPaddingKey = "ENGFLOW.padding"

type counter struct {
	n int
}

func (c *counter) Write(d []byte) (int, error) {
	n := len(d)
	c.n += n
	return n, nil
}

func hdrSize(hdr *Header) int64 {
	buf := counter{}
	tw := NewWriter(&buf)
	h := *hdr
	tw.WriteHeader(&h)
	// no need to call flush; WriteHeader always writes multiples of 512.
	return int64(buf.n)
}

func alignTo(hdr *Header, tarSize, bs int64) error {
	for {
		if (tarSize+hdrSize(hdr))%bs == 0 {
			return nil
		}
		if hdr.PAXRecords == nil {
			hdr.PAXRecords = map[string]string{}
		}
		hdr.PAXRecords[paxPaddingKey] = ""

		needPadding := bs - (tarSize+hdrSize(hdr))%bs
		if needPadding == 0 {
			return nil
		}

		h := *hdr
		h.PAXRecords = nil
		pureHeaderSz := hdrSize(&h)

		// Account for TypeXHeader which is gone because we
		// cleared PAXRecords
		pureHeaderSz += blockSize

		r, err := paxSpecialFile(hdr.PAXRecords)
		if err != nil {
			return err
		}

		missing := bs - (tarSize+(int64(len(r))+pureHeaderSz))%bs

		// Filling out exactly is tricky to get right, b/c the
		// size field in the PAX k/v encoding is variable
		// size. Just get to the middle of the tar block.
		missing -= blockSize / 2

		hdr.PAXRecords[paxPaddingKey] = strings.Repeat("x", int(missing))
		if needPadding := (bs - (tarSize+hdrSize(hdr))%bs) % bs; needPadding != 0 {
			return fmt.Errorf("giving up") // this shouldn't happen, but don't crash.
		}
	}
	return nil
}
