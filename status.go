package main

import (
	"errors"
	"fmt"

	"github.com/seletskiy/hierr"
)

type MirrorStatus struct {
	Name       string      `json:"name"`
	State      MirrorState `json:"state"`
	ModifyDate interface{} `json:"modify_date"`
}

type UpstreamStatus struct {
	Total          int            `json:"total"`
	Error          int            `json:"error"`
	ErrorPercent   float64        `json:"error_percent"`
	Success        int            `json:"success"`
	SuccessPercent float64        `json:"success_percent"`
	Slaves         []ServerStatus `json:"slaves"`
}
type ServerStatus struct {
	Address string         `json:"address,omitempty"`
	Error   error          `json:"error,omitempty"`
	Mirrors []MirrorStatus `json:"mirrors"`
}

type MasterServerStatus struct {
	Master   ServerStatus   `json:"master"`
	Upstream UpstreamStatus `json:"upstream"`
}

func (mirror MirrorStatus) HierarchicalError() string {
	return hierr.Push(mirror.Name,
		fmt.Sprintf("state: %s", mirror.State.String()),
		fmt.Sprintf("modify date: %v", mirror.ModifyDate),
	).Error()
}

func (upstream UpstreamStatus) HierarchicalError() string {
	header := fmt.Sprintf(
		"total: %v, success: %v (%.2f%%), error: %v (%.2f%%)",
		upstream.Total,
		upstream.Success, upstream.SuccessPercent,
		upstream.Error, upstream.ErrorPercent,
	)

	slaves := errors.New("slaves")
	for _, slave := range upstream.Slaves {
		slaves = hierr.Push(slaves, slave.HierarchicalError())
	}

	err := hierr.Push(header, slaves)

	return err.Error()
}

func (server ServerStatus) HierarchicalError() string {
	header := server.Address
	if header == "" {
		header = "upstream"
	}

	err := errors.New(header)
	if server.Error != nil {
		err = hierr.Push(
			"error", server.Error,
		)
	}

	mirrors := errors.New("mirrors")
	for _, mirror := range server.Mirrors {
		mirrors = hierr.Push(mirrors, mirror.HierarchicalError())
	}

	err = hierr.Push(err, mirrors)

	return err.Error()
}

func (master MasterServerStatus) HierarchicalError() string {
	return hierr.Push(
		"service status",
		hierr.Push("master", master.Master),
		master.Upstream,
	).Error()
}
