package netconf

import (
	"encoding/xml"
	"io"
)

const (
	msgDelim = "]]>]]>"
)

func WriteResponseWithOptions(response any, out io.Writer, includeMessageDelim bool, pretty bool) error {
	e := xml.NewEncoder(out)
	if pretty {
		e.Indent("", "  ")
	}
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

func WriteResponse(response any, out io.Writer) error {
	return WriteResponseWithOptions(response, out, false, false)
}
