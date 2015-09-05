package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type GitProxyConnection struct {
	mirrorStorageDir string
	mirrorStateTable *MirrorStateTable

	listenConn *net.TCPConn
	listenAddr *net.TCPAddr
	daemonAddr *net.TCPAddr
	id         int64

	daemonConn *net.TCPConn

	mirrorName string
	workers    *sync.WaitGroup
}

func (connection *GitProxyConnection) logf(
	format string, value ...interface{},
) {
	connection.log(fmt.Sprintf(format, value...))
}

func (connection *GitProxyConnection) log(value interface{}) {
	log.Printf("[git:%d] %s", connection.id, value)
}

func (connection *GitProxyConnection) Serve() {
	defer connection.listenConn.Close()

	buffer, err := connection.read(connection.listenConn)
	if err != nil {
		connection.log(err)
		return
	}

	err = connection.parseRequest(buffer)
	if err != nil {
		connection.logf("can't parse request: %s, refusing connection", err)
		return
	}

	err = connection.validateMirror()
	if err != nil {
		connection.logf("can't validate mirror: %s, refusing connection", err)
		return
	}

	connection.logf(
		"forwarding request with mirror '%s' to git daemon",
		connection.mirrorName,
	)

	daemonConn, err := net.DialTCP("tcp", nil, connection.daemonAddr)
	if err != nil {
		connection.logf("can't connect to git daemon: %s", err)
		return
	}

	connection.daemonConn = daemonConn
	defer connection.daemonConn.Close()

	// enabling Nagle's Algorithm
	connection.listenConn.SetNoDelay(true)
	connection.daemonConn.SetNoDelay(true)

	err = connection.write(connection.daemonConn, buffer)
	if err != nil {
		connection.log(err)
		return
	}

	connection.workers = &sync.WaitGroup{}
	connection.workers.Add(2)

	// bidirectioal transfer
	go connection.pipe(connection.daemonConn, connection.listenConn, false)
	go connection.pipe(connection.listenConn, connection.daemonConn, true)

	connection.workers.Wait()

	connection.logf("request successfully forwarded")
}

func (connection *GitProxyConnection) pipe(
	src, dst *net.TCPConn, isLocalToRemote bool,
) {
	defer func() {
		connection.workers.Done()
	}()

	for {
		buffer, err := connection.read(src)
		if err != nil {
			if err != io.EOF {
				connection.log(err)
			}
			return
		}

		err = connection.write(dst, buffer)
		if err != nil {
			if err != io.EOF {
				connection.log(err)
			}
			return
		}
	}
}

func (connection *GitProxyConnection) read(conn *net.TCPConn) ([]byte, error) {
	buffer := make([]byte, 0xffff)
	bytes, err := conn.Read(buffer)
	if err != nil {
		return nil, err
	}

	return buffer[:bytes], nil

}

func (connection *GitProxyConnection) write(
	conn *net.TCPConn, buffer []byte,
) error {
	_, err := conn.Write(buffer)
	return err
}

func (connection *GitProxyConnection) parseRequest(buffer []byte) error {
	// ignore first four bytes - it's "pkt-len" (packet length), which not need,
	// because buffer will be splitted by \00 byte
	request := bytes.Split(buffer[4:], []byte{0x00})
	if len(request) < 2 {
		return fmt.Errorf("protocol error: %s", buffer)
	}

	uploadPackPart := request[0]

	// there is additional space after 'git-upload-pack' for preventing
	// ambigious binary name like 'git-upload-pack-fake-do-something-evil'
	if !bytes.HasPrefix(uploadPackPart, []byte("git-upload-pack ")) {
		return fmt.Errorf(
			"protocol-error: buffer should contains 'git-upload-pack' %s",
			buffer,
		)
	}

	mirrorName := strings.TrimPrefix(string(uploadPackPart), "git-upload-pack")
	mirrorName = strings.Trim(mirrorName, " /")

	if mirrorName == "" {
		return fmt.Errorf("protocol error: mirror name can't be empty")
	}

	connection.mirrorName = mirrorName

	return nil
}

func (connection *GitProxyConnection) validateMirror() error {
	mirror, err := GetMirror(
		connection.mirrorStorageDir, connection.mirrorName,
	)
	if err != nil {
		return fmt.Errorf("can't find mirror: %s", err)
	}

	mirrorState := connection.mirrorStateTable.GetState(mirror.Name)

	if mirrorState != MirrorStateSuccess {
		err = mirror.Pull()
		if err != nil {
			connection.logf("can't pull mirror '%s': %s", mirror.Name, err)
			connection.mirrorStateTable.SetState(
				mirror.Name, MirrorStateFailed,
			)

			return fmt.Errorf(
				"mirror state is '%s', error occurred while pulling data: %s",
				mirrorState, err,
			)
		}

		connection.mirrorStateTable.SetState(mirror.Name, MirrorStateSuccess)
		mirrorState = MirrorStateSuccess
	}

	if mirrorState != MirrorStateSuccess {
		return fmt.Errorf(
			"mirror '%s' state is '%s'", connection.mirrorName, mirrorState,
		)
	}

	return nil
}
