package netconf

import (
	"encoding/xml"
	"io"
)

const (
	msgDelim = "]]>]]>"
)

func WriteResponse(response any, out io.Writer, includeMessageDelim bool) error {
	e := xml.NewEncoder(out)
	if err := e.Encode(response); err != nil {
		return err
	}
	if includeMessageDelim {
		_, err := out.Write([]byte(msgDelim))
		if err != nil {
			return err
		}
	}
	return nil
}
