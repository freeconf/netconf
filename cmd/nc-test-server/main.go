package main

import (
	"github.com/freeconf/netconf"
	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/estream"
	"github.com/freeconf/yang/fc"
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/source"
	"github.com/freeconf/yang/testdata/car"
)

// Start the car app with NETCONF Server support to test NETCONF clients
// against.  This is unliekly a very useful tool long term so this will eventually
// turn into an example I suspect.
//
// hostkey was generated with
//   ssh-keygen -t rsa -f host.key

func main() {
	fc.DebugLog(true)
	c := car.New()
	api := car.Manage(c)
	ypath := source.Any(
		source.Dir("../../yang"),
		restconf.InternalYPath,
		restconf.InternalIetfRfcYPath,
		car.YPath,
	)
	d := device.New(ypath)
	d.Add("car", api)
	streams := estream.NewService()

	// in NETCONF, you pre-register streams you want to support.
	streams.AddStream(estream.Stream{
		Name: "car:update",
		Open: func() (*node.Selection, error) {
			b, err := d.Browser("car")
			if err != nil {
				return nil, err
			}
			return b.Root().Find("update")
		},
	})
	netconf.NewServer(d, streams)
	chkerr(d.ApplyStartupConfigFile("startup.json"))
	select {}
}

func chkerr(err error) {
	if err != nil {
		panic(err)
	}
}
