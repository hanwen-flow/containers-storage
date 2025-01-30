//go:build !go1.10

package archive

import (
	tar "github.com/hanwen-flow/hacktar"
)

func copyPassHeader(hdr *tar.Header) {
}

func maybeTruncateHeaderModTime(hdr *tar.Header) {
}
