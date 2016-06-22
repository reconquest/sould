package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/zazab/zhash"
)

// Server used for handling HTTP requests.
type Server struct {
	config       zhash.Hash
	states       *MirrorStates
	httpResource *http.Client
	insecureMode bool
}

// NewServer creates a new instance of Server, sets specified
// config as server config, if config is invalid, than special error will be
// returned.
func NewServer(
	config zhash.Hash, states *MirrorStates, insecureMode bool,
) (*Server, error) {
	server := Server{
		states:       states,
		insecureMode: insecureMode,
	}

	server.httpResource = &http.Client{
		Transport: &http.Transport{
			Dial: server.NetDial,
		},
	}

	err := server.SetConfig(config)
	if err != nil {
		return nil, hierr.Errorf(err, "invalid configuration")
	}

	return &server, nil
}

// SetConfig from zhash.Hash instance which actually is representation of
// map[string]interface{}
// SetConfig validates specified configuration before using.
func (server *Server) SetConfig(config zhash.Hash) error {
	isMaster, err := config.GetBool("master")
	if err != nil && !zhash.IsNotFound(err) {
		return err
	}

	if isMaster {
		slaves, err := config.GetStringSlice("slaves")
		if err != nil && !zhash.IsNotFound(err) {
			return err
		}

		if len(slaves) == 0 {
			logger.Warning("slave servers is not specified")
		} else {
			_, err = config.GetInt("timeout")
			if err != nil {
				return err
			}
		}
	}

	_, err = config.GetString("storage")
	if err != nil {
		return err
	}

	_, err = config.GetString("http", "listen")
	if err != nil {
		return err
	}

	server.config = config

	return nil
}

// IsMaster will be true only if config consists of special master directive.
func (server *Server) IsMaster() bool {
	isMaster, _ := server.config.GetBool("master")

	return isMaster
}

// IsSlave is shortcut for "not is master" and returns opposite value of
// Server.IsMaster()
func (server Server) IsSlave() bool {
	return !server.IsMaster()
}

// GetRole returns string representation of server role basing on server
// configuration.
// Can be: master or slave.
func (server Server) GetRole() string {
	if server.IsMaster() {
		return "master"
	}

	return "slave"
}

// GetStorageDir where to place all repository mirrors, value will be read from
// config.
func (server *Server) GetStorageDir() string {
	storage, _ := server.config.GetString("storage")

	return storage
}

// GetListenAddress which will be used for listening http connections, value
// will be read from config.
func (server *Server) GetListenAddress() string {
	address, _ := server.config.GetString("http", "listen")

	return address
}

// GetTimeout for all http actions, value will be read from config.
func (server *Server) GetTimeout() int64 {
	timeout, _ := server.config.GetInt("timeout")

	return timeout
}

// GetServersUpstream of sould slave servers which should be defined in
// configuration file if server is master, value will be read from config.
func (server *Server) GetServersUpstream() ServersUpstream {
	hosts, _ := server.config.GetStringSlice("slaves")

	return NewServersUpstream(hosts)
}

// NetDial sets timeout for dialing specified address.
func (server *Server) NetDial(
	network, address string,
) (net.Conn, error) {
	timeout := time.Duration(
		int64(time.Millisecond) * server.GetTimeout(),
	)

	return net.DialTimeout(network, address, timeout)
}

// GetMirror returns existing mirror or creates new instance in storage
// directory.
func (server *Server) GetMirror(
	name string, origin string,
) (mirror Mirror, created bool, err error) {
	mirror, err = GetMirror(server.GetStorageDir(), name)
	if err != nil {
		if !os.IsNotExist(err) {
			return Mirror{}, false, err
		}

		mirror, err = CreateMirror(server.GetStorageDir(), name, origin)
		if err != nil {
			return Mirror{}, false, NewError(err, "can't create new mirror")
		}

		return mirror, true, nil
	}

	mirrorURL, err := mirror.GetURL()
	if err != nil {
		return mirror, false, NewError(err, "can't get mirror origin url")
	}

	if mirrorURL != origin {
		return mirror, false, fmt.Errorf(
			"mirror have different origin url (%s)",
			mirrorURL,
		)
	}

	return mirror, true, nil
}
