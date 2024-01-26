package netconf

import (
	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/estream"
	"github.com/freeconf/yang/fc"
)

type Server struct {
	Ver        string
	main       device.Device
	sessNum    int64
	sshHandler *SshHandler
	streams    *estream.Service
}

type SessionManager interface {
	NextSessionId() int64
	StreamService() *estream.Service
	HandleErr(err error)
}

func NewServer(d *device.Local, streams *estream.Service) *Server {
	s := &Server{
		main:    d,
		streams: streams,
	}
	s.sshHandler = NewSshHandler(s, d)

	if err := d.Add("fc-netconf", Api(s)); err != nil {
		panic(err)
	}
	if err := d.Add("ietf-subscribed-notifications", estream.Manage(streams)); err != nil {
		panic(err)
	}
	return s
}

func (s *Server) StreamService() *estream.Service {
	return s.streams
}

func (s *Server) HandleErr(err error) {
	fc.Err.Print(err)
}

func (s *Server) NextSessionId() int64 {
	s.sessNum++
	return s.sessNum
}
