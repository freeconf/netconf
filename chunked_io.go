package netconf

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

// NewChunkedRdr reads "chunked" xml messages according to RFC6242 so that individual xml
// messages can be sent over a ssh stream.
//
//	see https://datatracker.ietf.org/doc/html/rfc6242#section-4.1
//
// Chunked message framing example where xxx is replaced with NETCONF messages:
//
//	  #4
//	  xxxx
//	  #12
//	  xxxxxxxxxxxx
//	  ##
//
//	This has 2 chunks of sizes 4 and 12 bytes.
func NewChunkedRdr(in io.Reader) <-chan io.Reader {
	buf := make([]byte, 4096)
	rdrs := make(chan io.Reader)
	chunked := bufio.NewReader(in)
	go func() {
		rdr, wtr := io.Pipe()
		rdrs <- rdr
		for {
			// RFC max size is 4.3 billion (uint32)
			chunkSize, err := readChunkSize(chunked)
			if err == io.EOF {
				wtr.Close()
				close(rdrs)
				return
			}
			if chunkSize == 0 {
				// don't close reader
				wtr.Close()
				rdr, wtr = io.Pipe()
				rdrs <- rdr
				continue
			}
			remainingSize := chunkSize
			for remainingSize > 0 {
				grabSize := minInt64(chunkSize, int64(len(buf)))
				readSize, err := chunked.Read(buf[:grabSize])
				if err != nil {
					panic(err)
				}
				wtr.Write(buf[:readSize])
				remainingSize -= int64(readSize)
			}
		}
	}()
	return rdrs
}

func readChunkSize(r io.ByteReader) (int64, error) {
	d := make([]byte, 0, 8)
	lf := 0
	hash := 0
	pos := 0
	for {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		switch b {
		case '\n':
			lf++
			if lf == 2 {
				if hash == 2 {
					if len(d) != 0 {
						return 0, fmt.Errorf("illegal chunked delimitor")
					}
					return 0, nil
				}
				return strconv.ParseInt(string(d), 10, 64)
			}
		case '#':
			hash++
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			d = append(d, b)
		default:
			return 0, fmt.Errorf("illegal framing char %d", b)
		}
		pos++
	}
}

func minInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// chunkedWtr
type chunkedWtr struct {
	raw io.Writer
}

func (cw chunkedWtr) Close() error {
	_, err := fmt.Fprint(cw.raw, "\n##\n")
	// we don't close delegate, it stays open for next chunked wtr
	return err
}

func (cw chunkedWtr) Write(p []byte) (int, error) {
	_, err := fmt.Fprintf(cw.raw, "\n#%d\n", len(p))
	if err != nil {
		return 0, err
	}
	// check the math, unclear if this check will ever be nec. or vary depending
	// on underlying io
	wrote, err := cw.raw.Write(p)
	if err != nil {
		return wrote, err
	}
	if wrote != len(p) {
		return wrote, fmt.Errorf("only wrote %d of %d bytes", wrote, len(p))
	}
	return wrote, err
}

// NewChunkedWtr is the counterpart to NewMsgsRdr
func NewChunkedWtr(w io.Writer) io.WriteCloser {
	return chunkedWtr{raw: w}
}
