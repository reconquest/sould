package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/kovetskiy/toml"
	"github.com/pquerna/ffjson/ffjson"
	"github.com/seletskiy/hierr"
)

type resetter ServerStatus

func (*resetter) MarshalJSON() {}
func (*resetter) MarshalTOML() {}

var (
	_ json.Marshaler = ServerStatus{}
	_ toml.Marshaler = ServerStatus{}
)

func (status ServerStatus) MarshalJSON() ([]byte, error) {
	var (
		buffer bytes.Buffer
		data   []byte
		err    error
	)

	if status.Role == "master" {
		data, err = ffjson.Marshal(resetter(status))
	} else {
		status.Role = "slave"
		data, err = ffjson.Marshal(status.BasicServerStatus)
	}

	err = json.Indent(&buffer, data, "", ServerStatusResponseIndentation)
	if err != nil {
		logger.Errorf("can't indent json %s: %s", data, err)
		return data, nil
	}

	return buffer.Bytes(), nil
}

func (status ServerStatus) MarshalTOML() ([]byte, error) {
	var buffer bytes.Buffer
	var err error

	encoder := toml.NewEncoder(&buffer)
	encoder.Indent = ServerStatusResponseIndentation
	if status.Role == "master" {
		err = encoder.Encode(resetter(status))
	} else {
		status.Role = "slave"
		err = encoder.Encode(status.BasicServerStatus)
	}

	return buffer.Bytes(), err
}

func (status ServerStatus) MarshalHierarchical() []byte {
	var hierarchy error
	if status.Address != "" {
		hierarchy = hierr.Push(status.Address)
	} else {
		hierarchy = hierr.Push("status")
	}

	if status.Role != "master" {
		status.Role = "slave"
	}

	hierarchy = hierr.Push(
		hierarchy,
		fmt.Sprintf("role: %s", status.Role),
	)

	hierarchy = hierr.Push(
		hierarchy,
		fmt.Sprintf("total: %d", len(status.Mirrors)),
	)

	if status.HierarchicalError != "" {
		hierarchy = hierr.Push(
			hierarchy,
			hierr.Push("error", status.HierarchicalError),
		)
	}

	if len(status.Mirrors) > 0 {
		mirrors := errors.New("mirrors")
		for _, mirror := range status.Mirrors {
			mirrors = hierr.Push(mirrors, mirror.Hierarchical())
		}
		hierarchy = hierr.Push(hierarchy, mirrors)
	}

	if status.Role == "master" {
		hierarchy = hierr.Push(hierarchy, status.Upstream.Hierarchical())
	}

	return []byte(hierr.String(hierarchy))
}

func (status MirrorStatus) Hierarchical() string {
	hierarchy := hierr.Push(
		status.Name,
		fmt.Sprintf("state: %s", status.State),
	)

	if status.ModifyDate > 0 {
		hierarchy = hierr.Push(
			hierarchy,
			fmt.Sprintf("modify date: %v", status.ModifyDate),
		)
	}

	return hierr.String(hierarchy)
}

func (status UpstreamStatus) Hierarchical() string {
	hierarchy := hierr.Push(
		"upstream",
		"total: "+strconv.Itoa(status.Total),
		fmt.Sprintf(
			"success: %d (%.2f%%)",
			status.Success, status.SuccessPercent,
		),
		fmt.Sprintf(
			"error: %d (%.2f%%)",
			status.Error, status.ErrorPercent,
		),
	)

	if len(status.Slaves) > 0 {
		slaves := errors.New("slaves")
		for _, slave := range status.Slaves {
			slaves = hierr.Push(slaves, slave.MarshalHierarchical())
		}

		hierarchy = hierr.Push(hierarchy, slaves)
	}

	return hierr.String(hierarchy)
}
