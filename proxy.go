package main

import (
	"fmt"
	"net"

	"github.com/zazab/zhash"
)

// GitProxy is proxy server for git daemon.
type GitProxy struct {
	listenAddr  *net.TCPAddr
	daemonAddr  *net.TCPAddr
	listener    *net.TCPListener
	connections int64
	states      *MirrorStates
	storageDir  string
}

// NewGitProxy creates new instance of git proxy server using specified
// configuration.
func NewGitProxy(
	config zhash.Hash, states *MirrorStates,
) (*GitProxy, error) {
	proxy := &GitProxy{}

	err := proxy.SetConfig(config)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %s", err)
	}

	proxy.states = states

	return proxy, nil
}

// SetConfig validates,  sets given configuration and resolv listen and daemon
// addresses using new configuration variables.
func (proxy *GitProxy) SetConfig(config zhash.Hash) error {
	listenAddrString, err := config.GetString("git", "listen")
	if err != nil {
		return err
	}

	daemonAddrString, err := config.GetString("git", "daemon")
	if err != nil {
		return err
	}

	storageDir, err := config.GetString("storage")
	if err != nil {
		return err
	}

	listenAddr, err := net.ResolveTCPAddr("tcp", listenAddrString)
	if err != nil {
		return err
	}

	daemonAddr, err := net.ResolveTCPAddr("tcp", daemonAddrString)
	if err != nil {
		return err
	}

	proxy.storageDir = storageDir
	proxy.listenAddr = listenAddr
	proxy.daemonAddr = daemonAddr

	return nil
}

// Start creates new tcp listener and starts new thread for handling
// connections.
func (proxy *GitProxy) Start() error {
	var err error
	proxy.listener, err = net.ListenTCP("tcp", proxy.listenAddr)
	if err != nil {
		return err
	}

	go proxy.handle()

	return nil
}

// Stop closes listenning tcp connection.
func (proxy *GitProxy) Stop() {
	proxy.listener.Close()
}

func (proxy *GitProxy) handle() {
	for {
		client, err := proxy.listener.AcceptTCP()
		if err != nil {
			logger.Infof("proxy connection accept failed: %s", err)
			break
		}

		proxy.connections++

		connection := &GitProxyConnection{
			client:     client,
			daemonAddr: proxy.daemonAddr,
			id:         proxy.connections,
			states:     proxy.states,
			storageDir: proxy.storageDir,
		}

		go connection.Handle()
	}
}
