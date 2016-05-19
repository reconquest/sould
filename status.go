package main

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/seletskiy/hierr"
)

type MirrorStatus struct {
	Name       string      `json:"name" toml:"name"`
	State      MirrorState `json:"state" toml:"state"`
	ModifyDate interface{} `json:"modify_date" toml:"modify_date"`
}

type UpstreamStatus struct {
	Total          int            `json:"total" toml:"total"`
	Error          int            `json:"error" toml:"error"`
	ErrorPercent   float64        `json:"error_percent" toml:"error_percent"`
	Success        int            `json:"success" toml:"success"`
	SuccessPercent float64        `json:"success_percent" toml:"success_percent"`
	Slaves         []ServerStatus `json:"slaves,omitempty" toml:"slaves,omitempty"`
}

type ServerStatus struct {
	Role    string         `json:"role,omitempty" toml:"role,omitempty"`
	Address string         `json:"address,omitempty" toml:"address,omitempty"`
	Error   error          `json:"error,omitempty" toml:"error,omitempty"`
	Total   int            `json:"total" toml:"total"`
	Mirrors []MirrorStatus `json:"mirrors,omitempty" toml:"mirrors,omitempty"`
}

type MasterServerStatus struct {
	Role     string         `json:"role" toml:"role"`
	Server   ServerStatus   `json:"master" toml:"master"`
	Upstream UpstreamStatus `json:"upstream" toml:"upstream"`
}

func (mirror MirrorStatus) HierarchicalError() string {
	return hierr.Push(mirror.Name,
		fmt.Sprintf("state: %s", mirror.State.String()),
		fmt.Sprintf("modify date: %v", mirror.ModifyDate),
	).Error()
}

func (upstream UpstreamStatus) HierarchicalError() string {
	err := hierr.Push("upstream",
		"total: "+strconv.Itoa(upstream.Total),
		fmt.Sprintf(
			"success: %d (%.2f%%)",
			upstream.Success, upstream.SuccessPercent,
		),
		fmt.Sprintf(
			"error: %d (%.2f%%)",
			upstream.Error, upstream.ErrorPercent,
		),
	)

	if len(upstream.Slaves) > 0 {
		slaves := errors.New("slaves")
		for _, slave := range upstream.Slaves {
			slaves = hierr.Push(slaves, slave.HierarchicalError())
		}

		err = hierr.Push(err, slaves)
	}

	return err.Error()
}

func (server ServerStatus) HierarchicalError() string {
	var err error
	switch {
	case server.Address != "":
		err = errors.New(server.Address)

	case server.Role == "master":
		err = errors.New("master")

	case server.Role == "slave":
		err = hierr.Push(
			"status",
			"role: slave",
		)
	}

	if server.Error != nil {
		err = hierr.Push(
			"error", server.Error,
		)
	}

	err = hierr.Push(
		err,
		fmt.Sprintf("total: %d", len(server.Mirrors)),
	)

	if len(server.Mirrors) > 0 {
		mirrors := errors.New("mirrors")
		for _, mirror := range server.Mirrors {
			mirrors = hierr.Push(mirrors, mirror.HierarchicalError())
		}
		err = hierr.Push(err, mirrors)
	}

	return err.Error()
}

func (master MasterServerStatus) HierarchicalError() string {
	master.Server.Role = "master"

	return hierr.Push(
		"status",
		master.Server,
		master.Upstream,
	).Error()
}
