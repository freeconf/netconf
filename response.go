package netconf

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
)

const (
	msgDelim = "]]>]]>"
)

func WriteResponse(response any, out io.Writer, includeMessageDelim bool) error {
	var buf bytes.Buffer
	e := xml.NewEncoder(&buf)
	if err := e.Encode(response); err != nil {
		return err
	}
	if includeMessageDelim {
		_, err := buf.Write([]byte(msgDelim))
		if err != nil {
			return err
		}
	}
	_, err := out.Write(buf.Bytes())
	fmt.Printf("sent response %v\n", buf.String())
	return err
}
