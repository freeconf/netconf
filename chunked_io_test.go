package netconf

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/freeconf/yang/fc"
)

func TestChunkedRdr(t *testing.T) {
	msg := `
#4
1234
#10
1234567890
##

#1
1
##
`
	rdrs := NewChunkedRdr(strings.NewReader(msg))
	msg1, err := io.ReadAll(<-rdrs)
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, `12341234567890`, string(msg1))
	msg2, err := io.ReadAll(<-rdrs)
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, `1`, string(msg2))
	msg3, err := io.ReadAll(<-rdrs)
	fc.AssertEqual(t, nil, err)

	// unfortunately artifact of underlying implementation is that
	// there will always be an empty reader just before closing.
	fc.AssertEqual(t, ``, string(msg3))

	fc.AssertEqual(t, nil, <-rdrs)
}

func TestChunkedWtr(t *testing.T) {
	var buf bytes.Buffer
	msg := NewChunkedWtr(&buf)
	n, err := msg.Write([]byte("1234567890"))
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, 10, n)
	fc.AssertEqual(t, nil, msg.Close())

	msg = NewChunkedWtr(&buf)
	n, err = msg.Write([]byte("1234"))
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, 4, n)
	n, err = msg.Write([]byte("1"))
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, 1, n)
	fc.AssertEqual(t, nil, msg.Close())

	expected := `
#10
1234567890
##

#4
1234
#1
1
##
`
	fc.AssertEqual(t, expected, buf.String())
}
