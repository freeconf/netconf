package netconf

import (
	"testing"

	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/estream"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/nodeutil"
	"github.com/freeconf/yang/source"
)

func TestServer(t *testing.T) {
	ypath := source.Any(
		restconf.InternalYPath,
		source.Dir("./yang"),
	)
	streams := estream.NewService()
	d := device.New(ypath)
	s := NewServer(d, streams)
	d.Add("fc-netconf", Api(s))
	b, err := d.Browser("fc-netconf")
	fc.RequireEqual(t, nil, err)
	err = b.Root().UpdateFrom(nodeutil.ReadJSON(`
	{
		"ssh": {
			"options" : {
				"port" : "127.0.0.1:9000",
				"hostKeyFile" : "testdata/host.key"
			}
		}
	}
	`))
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, true, s.sshHandler.Status().Running)
	// configure again to test service live re-config
	err = b.Root().UpdateFrom(nodeutil.ReadJSON(`
	{
		"ssh": {
			"options" : {
				"port" : "127.0.0.1:9001",
				"hostKeyFile" : "testdata/host.key"
			}
		}
	}
	`))
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, true, s.sshHandler.Status().Running)
}
