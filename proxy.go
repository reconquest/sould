package main

import (
	"log"
	"net"

	"github.com/zazab/zhash"
)

type GitProxy struct {
	listenAddr       *net.TCPAddr
	daemonAddr       *net.TCPAddr
	listener         *net.TCPListener
	connections      int64
	mirrorStateTable *MirrorStateTable
	mirrorStorageDir string
}

func NewGitProxy(
	config zhash.Hash, mirrorStateTable *MirrorStateTable,
) (*GitProxy, error) {
	var err error

	proxy := &GitProxy{}
	proxy.SetConfig(config)

	proxy.mirrorStateTable = mirrorStateTable

	proxy.listener, err = net.ListenTCP("tcp", proxy.listenAddr)
	if err != nil {
		return nil, err
	}

	return proxy, nil
}

func (proxy *GitProxy) SetConfig(config zhash.Hash) error {
	listenAddrString, err := config.GetString("git", "listen")
	if err != nil {
		return err
	}

	daemonAddrString, err := config.GetString("git", "daemon")
	if err != nil {
		return err
	}

	mirrorStorageDir, err := config.GetString("storage")
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

	proxy.mirrorStorageDir = mirrorStorageDir
	proxy.listenAddr = listenAddr
	proxy.daemonAddr = daemonAddr

	return nil
}

func (proxy *GitProxy) Start() {
	go proxy.loop()
}

func (proxy *GitProxy) Stop() {
	proxy.listener.Close()
}

func (proxy *GitProxy) loop() {
	for {
		tcpConn, err := proxy.listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			break
		}

		proxy.connections++

		connection := &GitProxyConnection{
			listenConn:       tcpConn,
			listenAddr:       proxy.listenAddr,
			daemonAddr:       proxy.daemonAddr,
			id:               proxy.connections,
			mirrorStateTable: proxy.mirrorStateTable,
			mirrorStorageDir: proxy.mirrorStorageDir,
		}

		go connection.Serve()
	}
}
