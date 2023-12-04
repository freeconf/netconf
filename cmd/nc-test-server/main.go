package main

import (
	"github.com/freeconf/netconf"
	"github.com/freeconf/restconf"
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/testdata"
	"github.com/freeconf/yang/source"
)

// Start the cara app with NETCONF Server support to test NETCONF clients
// against.  This is unliekly a very useful tool long term so this will eventually
// turn into an example I suspect.
//
// hostkey was generated with
//   ssh-keygen -t rsa -f host.key

func main() {
	c := testdata.New()
	api := testdata.Manage(c)
	ypath := source.Any(
		source.Path(".:../../yang"),
		restconf.InternalYPath,
	)
	d := device.New(ypath)
	d.Add("car", api)
	netconf.NewServer(d)
	chkerr(d.ApplyStartupConfigFile("startup.json"))
	select {}
}

func chkerr(err error) {
	if err != nil {
		panic(err)
	}
}
