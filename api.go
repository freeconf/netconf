package netconf

import (
	"github.com/freeconf/yang/node"
	"github.com/freeconf/yang/nodeutil"
)

type api struct{}

func Api(s *Server) node.Node {
	var api api
	return &nodeutil.Basic{
		OnChild: func(r node.ChildRequest) (child node.Node, err error) {
			switch r.Meta.Ident() {
			case "ssh":
				return api.ssh(s.sshHandler), nil
			}
			return nil, nil
		},
	}
}

func (api api) ssh(s *SshHandler) node.Node {
	return &nodeutil.Node{Object: s,
		OnChild: func(n *nodeutil.Node, r node.ChildRequest) (child node.Node, err error) {
			switch r.Meta.Ident() {
			case "options":
				return api.sshOptions(s), nil
			case "status":
				s := s.Status()
				return n.New(&s), nil
			}
			return n.DoChild(r)
		},
	}
}

func (api api) sshOptions(s *SshHandler) node.Node {
	var opts = s.Options()
	return &nodeutil.Node{Object: &opts,
		OnEndEdit: func(n *nodeutil.Node, r node.NodeRequest) error {
			return s.Apply(opts)
		},
	}
}
