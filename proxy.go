package main

import (
	"fmt"
	"log"
	"net"
)

type GitProxy struct {
	localAddr     string
	remoteAddr    string
	netLocalAddr  *net.TCPAddr
	netRemoteAddr *net.TCPAddr
	listener      *net.TCPListener
	pipeErrors    chan error
	connections   int64
}

func NewGitProxy(localAddr, remoteAddr string) (*GitProxy, error) {
	netLocalAddr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		return nil, err
	}

	netRemoteAddr, err := net.ResolveTCPAddr("tcp", remoteAddr)
	if err != nil {
		return nil, err
	}

	listener, err := net.ListenTCP("tcp", netLocalAddr)
	if err != nil {
		return nil, err
	}

	return &GitProxy{
		localAddr:     localAddr,
		remoteAddr:    remoteAddr,
		netLocalAddr:  netLocalAddr,
		netRemoteAddr: netRemoteAddr,
		listener:      listener,
	}, nil
}

func (proxy *GitProxy) RunLoop() {
	go proxy.loop()
}

func (proxy *GitProxy) loop() {
	for {
		tcpConn, err := proxy.listener.AcceptTCP()
		if err != nil {
			log.Println(err)
			continue
		}
		fmt.Printf("XXXXXX proxy.go:55: connection accepeted\n")

		proxy.connections++

		connection := &GitProxyConnection{
			localConn:  tcpConn,
			localAddr:  proxy.netLocalAddr,
			remoteAddr: proxy.netRemoteAddr,
			id:         proxy.connections,
		}

		go connection.Serve()
	}
}
