package main

type MirrorStatus struct {
	Name       string `json:"name" toml:"name"`
	State      string `json:"state" toml:"state"`
	ModifyDate int64  `json:"modify_date,omitempty,omitzero" toml:"modify_date,omitempty,omitzero"`
}

type UpstreamStatus struct {
	Total          int            `json:"total" toml:"total"`
	Error          int            `json:"error" toml:"error"`
	ErrorPercent   float64        `json:"error_percent" toml:"error_percent"`
	Success        int            `json:"success" toml:"success"`
	SuccessPercent float64        `json:"success_percent" toml:"success_percent"`
	Slaves         []ServerStatus `json:"slaves,omitempty" toml:"slaves,omitempty"`
}

type BasicServerStatus struct {
	Address           string         `json:"address,omitempty" toml:"address,omitempty"`
	Role              string         `json:"role,omitempty" toml:"role,omitempty"`
	Total             int            `json:"total" toml:"total"`
	Error             string         `json:"error,omitempty" toml:"error,omitempty"`
	HierarchicalError string         `json:"heararchical_error,omitempty" toml:"heararchical_error,omitempty"`
	Mirrors           []MirrorStatus `json:"mirrors,omitempty" toml:"mirrors,omitempty"`
}

type ServerStatus struct {
	BasicServerStatus
	Upstream UpstreamStatus `json:"upstream" toml:"upstream"`
}
