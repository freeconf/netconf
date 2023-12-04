package netconf

import (
	"github.com/freeconf/restconf/device"
)

type Server struct {
	Ver        string
	main       device.Device
	sessNum    int64
	sshHandler *SshHandler
}

type SessionManager interface {
	NextSessionId() int64
	HandleErr(err error)
}

func NewServer(d *device.Local) *Server {
	s := &Server{
		main: d,
	}
	s.sshHandler = NewSshHandler(s, d)

	if err := d.Add("fc-netconf", Api(s)); err != nil {
		panic(err)
	}
	return s
}

func (s *Server) HandleErr(err error) {
	panic(err)
}

func (s *Server) NextSessionId() int64 {
	s.sessNum++
	return s.sessNum
}
