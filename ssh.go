package netconf

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/freeconf/restconf/device"
	"github.com/freeconf/restconf/secure"
	"github.com/freeconf/yang/fc"
	"golang.org/x/crypto/ssh"
)

type SshHandler struct {
	opts                 SshOptions
	Auth                 secure.Auth
	listener             net.Listener
	hostPrivateKeySigner ssh.Signer
	host                 SessionManager
	dev                  device.Device
}

type SshOptions struct {
	Port        string
	HostKeyFile string
}

func NewSshHandler(h SessionManager, dev device.Device) *SshHandler {
	return &SshHandler{
		host: h,
		dev:  dev,
	}
}

func (s *SshHandler) Options() SshOptions {
	return s.opts
}

func (s *SshHandler) Apply(opts SshOptions) error {
	if s.opts == opts {
		return nil
	}
	if opts.Port == "" {
		return fmt.Errorf("missing port")
	}
	if opts.HostKeyFile == "" {
		return fmt.Errorf("invalid hosts key file")
	}
	hostPrivateKey, err := os.ReadFile(opts.HostKeyFile)
	if err != nil {
		return fmt.Errorf("could not read hosts key file %s. %w", opts.HostKeyFile, err)
	}
	s.hostPrivateKeySigner, err = ssh.ParsePrivateKey(hostPrivateKey)
	if err != nil {
		panic(err)
	}
	s.opts = opts
	if s.listener != nil {
		s.listener.Close()
	}
	return s.start()
}

type SshStatus struct {
	Running bool
}

func (s *SshHandler) Status() SshStatus {
	return SshStatus{
		Running: s.listener != nil,
	}
}

func (s *SshHandler) keyAuth(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	log.Println(conn.RemoteAddr(), "authenticate with", key.Type())
	return nil, nil
}

func (s *SshHandler) handleNewChannels(conn *ssh.ServerConn, newChannelRequests <-chan ssh.NewChannel) {
	defer conn.Close()
	for req := range newChannelRequests {
		go s.handleConn(conn, req)
	}
}

func (s *SshHandler) handleConn(conn *ssh.ServerConn, c ssh.NewChannel) {
	fmt.Println("got connection, waiting for message...")
	if c.ChannelType() != "session" {
		c.Reject(ssh.Prohibited, "channel type is not a session")
		return
	}
	ch, reqs, err := c.Accept()
	if err != nil {
		log.Println("fail to accept channel request", err)
		return
	}
	sess := NewSession(s.host, s.dev, ch, ch)
	go func(in <-chan *ssh.Request) {
		defer ch.Close()
		for req := range reqs {
			fmt.Printf("req, type=%s, payload=%v\n", req.Type, req.Payload)
			switch req.Type {
			case "subsystem":
				if len(req.Payload) >= 4 {
					if bytes.Equal(req.Payload[4:], []byte("netconf")) {

						// sending hello shouldn't wait to recieve message from client
						// https://datatracker.ietf.org/doc/html/rfc6242#section-3.1
						go func() {
							if serr := WriteResponse(sess.Hello(), ch, true); serr != nil {
								s.host.HandleErr(serr)
							}
						}()

						if req.WantReply {
							req.Reply(true, []byte{})
						}
						if err := sess.readMessages(); err != nil {
							s.host.HandleErr(err)
							break
						}
						continue
					}
				}
				fmt.Printf("unexpected message")
				req.Reply(false, nil)
			default:
				c.Reject(ssh.Prohibited, "channel subtype is not supported")
				return
			}
		}
	}(reqs)
}

func (s *SshHandler) rejectGlobalRequests(requests <-chan *ssh.Request) {
	for request := range requests {
		if request.WantReply {
			_ = request.Reply(false, []byte("request type not supported"))
		}
	}
}

func (s *SshHandler) start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.opts.Port)
	if err != nil {
		return err
	}
	config := ssh.ServerConfig{
		PublicKeyCallback: s.keyAuth,
	}
	config.AddHostKey(s.hostPrivateKeySigner)

	go func() {
		defer s.listener.Close()
		for {
			c, err := s.listener.Accept()
			if err != nil {
				if x, ok := err.(*net.OpError); ok && x.Op == "accept" {
					fc.Info.Print("graceful shutdown")
					return
				}
				s.host.HandleErr(err)
			}
			sshConn, chans, globalRequests, err := ssh.NewServerConn(c, &config)
			if err != nil {
				s.host.HandleErr(err)
			} else {
				go s.handleNewChannels(sshConn, chans)
				go s.rejectGlobalRequests(globalRequests)
			}
		}
	}()

	return nil
}
