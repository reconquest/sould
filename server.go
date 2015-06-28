package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/zazab/zhash"
)

type MirrorServer struct {
	config       zhash.Hash
	stateTable   MirrorStateTable
	httpClient   *http.Client
	unsecureMode bool
}

func NewMirrorServer(
	config zhash.Hash, table MirrorStateTable, unsecureMode bool,
) (*MirrorServer, error) {
	server := MirrorServer{
		stateTable:   table,
		unsecureMode: unsecureMode,
	}

	server.httpClient = &http.Client{
		Transport: &http.Transport{
			Dial: server.NetDial,
		},
	}

	err := server.SetConfig(config)
	if err != nil {
		return nil, err
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
			log.Println("slave servers directive is empty or not defined")
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

	_, err = config.GetString("listen")
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

func (server *MirrorServer) GetStorageDir() string {
	storage, _ := server.config.GetString("storage")

	return storage
}

func (server *MirrorServer) GetListenAddress() string {
	address, _ := server.config.GetString("listen")

	return address
}

func (server *MirrorServer) GetTimeout() int64 {
	timeout, _ := server.config.GetInt("timeout")

	return timeout
}

func (server *MirrorServer) GetSlaves() MirrorSlaves {
	hosts, _ := server.config.GetStringSlice("slaves")

	return NewMirrorSlaves(hosts)
}

func (server *MirrorServer) NetDial(
	network, address string,
) (net.Conn, error) {
	timeout := time.Duration(
		int64(time.Microsecond) * server.GetTimeout(),
	)

	return net.DialTimeout(network, address, timeout)
}
