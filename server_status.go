package main

// MirrorStatus represents status of mirror.
type MirrorStatus struct {
	// Name of mirror
	Name string `json:"name" toml:"name"`

	// State is string representation of mirror from state table.
	State string `json:"state" toml:"state"`

	// ModifyDate in unix timestamp format.
	ModifyDate int64 `json:"modify_date,omitempty,omitzero" toml:"modify_date,omitempty,omitzero"`
}

// UpstreamStatus represents status of slave group.
type UpstreamStatus struct {
	// Total represents amount of slaves.
	Total int `json:"total" toml:"total"`

	// Error represents amount of slaves which unavailable.
	Error int `json:"error" toml:"error"`

	// ErrorPercent represents percent of error slaves.
	ErrorPercent float64 `json:"error_percent" toml:"error_percent"`

	// Success represents amount of slaves which available and responses.
	Success int `json:"success" toml:"success"`

	// SuccessPercent represents percent of success slaves.
	SuccessPercent float64 `json:"success_percent" toml:"success_percent"`

	// Slaves represents status of all slaves and their statuses.
	Slaves []ServerStatus `json:"slaves,omitempty" toml:"slaves,omitempty"`
}

// BasicServerStatus represents server status which should have all servers
// independently of server role.
type BasicServerStatus struct {
	// Address is a hostname or ip address.
	Address string `json:"address,omitempty" toml:"address,omitempty"`

	// Role can be slave or master.
	Role string `json:"role,omitempty" toml:"role,omitempty"`

	// Total is amount of mirrors on server.
	Total int `json:"total" toml:"total"`

	// Error is some server error which occurred during serving status request.
	Error string `json:"error,omitempty" toml:"error,omitempty"`

	// Error is hierarchical representation of Error field.
	HierarchicalError string `json:"hierarchical_error,omitempty" toml:"hierarchical_error,omitempty"`

	// Mirrors is slice of all mirrors and their statuses.
	Mirrors []MirrorStatus `json:"mirrors,omitempty" toml:"mirrors,omitempty"`
}

// ServerStatus represents status of server which extends BasicServerStatus and
// can propagate request to other servers.
type ServerStatus struct {
	BasicServerStatus
	// Upstream represents status of slave group.
	Upstream UpstreamStatus `json:"upstream" toml:"upstream"`
}
