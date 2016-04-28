package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/kovetskiy/lorg"
)

// GitProxyConnections is representation of TCP connection with sould git proxy
// and local git daemon which runs in storage directory.
type GitProxyConnection struct {
	id         int64
	client     *net.TCPConn
	daemon     *net.TCPConn
	daemonAddr *net.TCPAddr
	mirrorName string
	storageDir string
	states     *MirrorStates
	logger     lorg.Logger
}

func (connection *GitProxyConnection) bootstrap() {
	connection.logger = NewPrefixedLogger(
		fmt.Sprintf("(git:%d)", connection.id),
	)

	connection.logger.Infof(
		"remote_addr = %s",
		connection.client.RemoteAddr(),
	)
}

// Handle new tcp connection between client and git daemon.
func (connection *GitProxyConnection) Handle() {
	defer func() {
		err := recover()
		if err != nil {
			// using application-wide logger for preventing extra panics by
			// connection.logger
			logger.Errorf(
				"(git:%d) PANIC! %s\n%s",
				connection.id, err, stack(),
			)
		}
	}()

	connection.bootstrap()

	defer func() {
		connection.logger.Error("closing connection")

		err := connection.client.Close()
		if err != nil {
			connection.logger.Error(err)
		}
	}()

	err := connection.serve()
	if err != nil {
		connection.logger.Error(err)
	}
}

func (connection *GitProxyConnection) serve() error {
	buffer, err := connection.read(connection.client)
	if err != nil {
		return fmt.Errorf("can't read request: %s", err)
	}

	err = connection.parseRequest(buffer)
	if err != nil {
		return fmt.Errorf("can't parse request: %s", err)
	}

	err = connection.validateMirror()
	if err != nil {
		return fmt.Errorf("can't validate mirror: %s", err)
	}

	connection.logger.Infof(
		"serving request with mirror %s",
		connection.mirrorName,
	)

	daemon, err := net.DialTCP("tcp", nil, connection.daemonAddr)
	if err != nil {
		return fmt.Errorf(
			"can't connect to git daemon %s: %s",
			connection.daemonAddr, err,
		)
	}

	connection.daemon = daemon
	defer connection.daemon.Close()

	// enabling Nagle's Algorithm
	err = connection.client.SetNoDelay(true)
	if err != nil {
		return fmt.Errorf("can't set no delay for client: %s", err)
	}

	err = connection.daemon.SetNoDelay(true)
	if err != nil {
		return fmt.Errorf("can't set no delay for daemon: %s", err)
	}

	err = connection.write(connection.daemon, buffer)
	if err != nil {
		return fmt.Errorf("can't write buffer: %s", err)
	}

	connection.pipe()

	return nil
}

func (connection *GitProxyConnection) pipe() {
	pipers := &sync.WaitGroup{}
	pipers.Add(2)

	go func() {
		defer pipers.Done()

		_, err := io.Copy(connection.daemon, connection.client)
		if err != nil {
			connection.logger.Errorf("can't proxy daemon -> client: %s", err)
		}
	}()

	go func() {
		defer pipers.Done()

		_, err := io.Copy(connection.client, connection.daemon)
		if err != nil {
			connection.logger.Errorf("can't proxy client -> daemon: %s", err)
		}
	}()

	pipers.Wait()
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
		return errors.New(
			"protocol error: unexpected packet",
		)
	}

	uploadPackPart := request[0]

	// there is additional space after 'git-upload-pack' for preventing
	// ambigious binary name like 'git-upload-pack-fake-do-something-evil'
	if !bytes.HasPrefix(uploadPackPart, []byte("git-upload-pack ")) {
		return errors.New(
			"protocol error: buffer should contains 'git-upload-pack'",
		)
	}

	mirrorName := strings.TrimPrefix(string(uploadPackPart), "git-upload-pack")
	mirrorName = strings.Trim(mirrorName, " /")

	if mirrorName == "" {
		return errors.New("protocol error: mirror name can't be empty")
	}

	connection.mirrorName = mirrorName

	return nil
}

func (connection *GitProxyConnection) validateMirror() error {
	mirror, err := GetMirror(
		connection.storageDir, connection.mirrorName,
	)
	if err != nil {
		return fmt.Errorf(
			"can't get mirror %s: %s", connection.mirrorName, err,
		)
	}

	state := connection.states.GetState(mirror.Name)

	fmt.Printf("XXXXXX proxy_conn.go:206: state: %#v\n", state.String())
	if state == MirrorStateUnknown || state == MirrorStateError {
		connection.states.SetState(mirror.Name, MirrorStateProcessing)

		connection.logger.Warning("XXXXX")
		connection.logger.Infof("fetching mirror %s changeset", mirror.String())

		err = mirror.Fetch()
		if err != nil {
			connection.states.SetState(
				mirror.Name, MirrorStateError,
			)

			return fmt.Errorf(
				"mirror state is '%s', error occurred while pulling data: %s",
				state, err,
			)
		}

		connection.logger.Infof("mirror %s changeset fetched", mirror.String())

		connection.states.SetState(mirror.Name, MirrorStateSuccess)
	}

	return nil
}
