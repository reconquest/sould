package main

import (
	"net"
	"net/http"
	"time"

	"github.com/seletskiy/hierr"
	"github.com/zazab/zhash"
)

type MirrorServer struct {
	config       zhash.Hash
	states       *MirrorStates
	httpResource *http.Client
	insecureMode bool
}

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

func (server *MirrorServer) IsMaster() bool {
	isMaster, _ := server.config.GetBool("master")

	return isMaster
}

func (server MirrorServer) IsSlave() bool {
	return !server.IsMaster()
}

func (server *MirrorServer) GetStorageDir() string {
	storage, _ := server.config.GetString("storage")

	return storage
}

func (server *MirrorServer) GetListenAddress() string {
	address, _ := server.config.GetString("http", "listen")

	return address
}

func (server *MirrorServer) GetTimeout() int64 {
	timeout, _ := server.config.GetInt("timeout")

	return timeout
}

func (server *MirrorServer) GetMirrorUpstream() MirrorUpstream {
	hosts, _ := server.config.GetStringSlice("slaves")

	return NewMirrorUpstream(hosts)
}

func (server *MirrorServer) NetDial(
	network, address string,
) (net.Conn, error) {
	timeout := time.Duration(
		int64(time.Millisecond) * server.GetTimeout(),
	)

	return net.DialTimeout(network, address, timeout)
}
