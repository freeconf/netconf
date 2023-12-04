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
	_, err := out.Write([]byte(msgDelim))
	return err
}

/*
type ResponseWriter struct {
	Out io.Writer
}

func (r ResponseWriter) sendHello(sesId int64) error {
	_, err := fmt.Fprintf(r.Out, `<hello xmlns="urn:ietf:params:xml:ns:netconf:base:1.0">
			<capabilities>
			<capability>
				urn:ietf:params:netconf:base:1.1
			</capability>
			</capabilities>
			<session-id>%d</session-id>
		</hello>]]>]]>`, sesId)
	if err != nil {
		return err
	}
	return nil
}

func (resp ResponseWriter) sendRpcOk(r *RpcMsg) error {
	if err := resp.openRpcReply(r); err != nil {
		return err
	}
	return resp.closeTag("rpc-reply")
}

func (resp ResponseWriter) openRpcReplyContent(rpc *RpcMsg, content string) error {
	if err := resp.openRpcReply(rpc); err != nil {
		return err
	}
	_, err := fmt.Fprintf(resp.Out, "<%s>", content)
	return err
}

func (resp ResponseWriter) closeRpcReplyContent(rpc *RpcMsg, content string) error {
	if err := resp.closeTag(content); err != nil {
		return err
	}
	return resp.closeTag("rpc-reply")
}

func (resp ResponseWriter) openRpcReply(rpc *RpcMsg) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "<rpc-reply")
	for _, a := range rpc.Attrs {
		fmt.Fprintf(&buf, ` %s="%s"`, a.Name, a.Value)
	}
	fmt.Fprintf(&buf, `>`)
	_, err := resp.Out.Write(buf.Bytes())
	return err
}

func (resp ResponseWriter) closeTag(tag string) error {
	fmt.Fprintf(resp.Out, `</%s>`, tag)
	return nil
}
*/
