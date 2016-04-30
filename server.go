package main

import (
	"net"
	"net/http"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/zazab/zhash"
)

// MirrorServer used for handling HTTP requests.
type MirrorServer struct {
	config       zhash.Hash
	states       *MirrorStates
	httpResource *http.Client
	insecureMode bool
}

// NewMirrorServer creates a new instance of MirrorServer, sets specified
// config as server config, if config is invalid, than special error will be
// returned.
func NewMirrorServer(
	config zhash.Hash, states *MirrorStates, insecureMode bool,
) (*MirrorServer, error) {
	server := MirrorServer{
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
func (server *MirrorServer) SetConfig(config zhash.Hash) error {
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
func (server *MirrorServer) IsMaster() bool {
	isMaster, _ := server.config.GetBool("master")

	return isMaster
}

// IsSlave is shortcut for "not is master" and returns opposite value of
// MirrorServer.IsMaster()
func (server MirrorServer) IsSlave() bool {
	return !server.IsMaster()
}

// GetStorageDir where to place all repository mirrors, value will be read from
// config.
func (server *MirrorServer) GetStorageDir() string {
	storage, _ := server.config.GetString("storage")

	return storage
}

// GetListenAddress which will be used for listening http connections, value
// will be read from config.
func (server *MirrorServer) GetListenAddress() string {
	address, _ := server.config.GetString("http", "listen")

	return address
}

// GetTimeout for all http actions, value will be read from config.
func (server *MirrorServer) GetTimeout() int64 {
	timeout, _ := server.config.GetInt("timeout")

	return timeout
}

// GetMirrorUpstream of sould slave servers which should be defined in
// configuration file if server is master, value will be read from config.
func (server *MirrorServer) GetMirrorUpstream() MirrorUpstream {
	hosts, _ := server.config.GetStringSlice("slaves")

	return NewMirrorUpstream(hosts)
}

// NetDial sets timeout for dialing specified address.
func (server *MirrorServer) NetDial(
	network, address string,
) (net.Conn, error) {
	timeout := time.Duration(
		int64(time.Millisecond) * server.GetTimeout(),
	)

	return net.DialTimeout(network, address, timeout)
}
