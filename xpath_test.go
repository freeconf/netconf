package netconf

import (
	"testing"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/testdata"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/source"
)

func TestCompileXPath(t *testing.T) {
	f := &RpcFilter{
		shortcodes: map[string]string{
			"cc": "c",
		},
		Type:   "xpath",
		Select: "cc:car/cc:tire",
	}
	ypath := source.Dir("./testdata/yang")
	d := device.New(ypath)
	c := testdata.New()
	api := testdata.Manage(c)
	d.Add("car", api)
	path, err := f.CompileXPath(d)
	fc.AssertEqual(t, nil, err)
	fc.AssertEqual(t, true, path != nil)
}
