package main

import (
	"fmt"
	"log"
	"net"
)

type GitProxyConnection struct {
	localConn  *net.TCPConn
	localAddr  *net.TCPAddr
	remoteAddr *net.TCPAddr
	id         int64

	remoteConn *net.TCPConn

	pipeErrors chan error
	pipeDone   chan bool
}

func (connection *GitProxyConnection) Serve() {
	defer connection.localConn.Close()

	remoteConn, err := net.DialTCP("tcp", nil, connection.remoteAddr)
	if err != nil {
		log.Println(err)
		return
	}

	connection.remoteConn = remoteConn
	defer connection.remoteConn.Close()

	// enabling Nagle's Algorithm
	connection.localConn.SetNoDelay(true)
	connection.remoteConn.SetNoDelay(true)

	localBuffer, err := connection.read(connection.localConn)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Printf("XXXXXX proxy_conn.go:49: localBuffer: %#v\n", localBuffer)

	err = connection.write(connection.remoteConn, localBuffer)
	if err != nil {
		log.Println(err)
		return
	}

	// bidirectioal transfer
	go connection.pipe(connection.remoteConn, connection.localConn, false)
	go connection.pipe(connection.localConn, connection.remoteConn, true)

	connection.workers.Wait()
}

func (connection *GitProxyConnection) pipe(
	src, dst *net.TCPConn, isLocalToRemote bool,
) {
	for {
		buffer, err := connection.read(src)
		if err != nil {
			log.Println(err)
			return
		}

		err = connection.write(dst, buffer)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func (connection *GitProxyConnection) read(src *net.TCPConn) ([]byte, error) {
	buffer := make([]byte, 0xffff)
	bytes, err := src.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:bytes], nil

}

func (connection *GitProxyConnection) write(dst *net.TCPConn, buffer []byte) error {
	_, err := dst.Write(buffer)
	return err
}

func (connection GitProxyConnection) error(err error) {
	pipeErrors <- err
}
