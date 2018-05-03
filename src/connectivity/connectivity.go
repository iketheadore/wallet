package connectivity

// Status represents a connection status.
type Status int

// Valid determines whether the status in question is valid.
func (s Status) Valid() bool {
	for _, si := range statuses {
		if si.Code == s {
			return true
		}
	}
	return false
}

const (
	Connected Status = iota
	Disconnected
	Connecting
	Unknown
)

// StatusInfo stores information about the specified status.
type StatusInfo struct {
	Code Status `json:"code"`
	Name string `json:"name"`
	Desc string `json:"description"`
}

// valid statuses.
var statuses = []StatusInfo{
	{
		Code: Connected,
		Name: "CONNECTED",
		Desc: "Connection is established.",
	},
	{
		Code: Disconnected,
		Name: "DISCONNECTED",
		Desc: "Connection is not established.",
	},
	{
		Code: Connecting,
		Name: "CONNECTING",
		Desc: "Attempting to connect.",
	},
	{
		Code: Unknown,
		Name: "UNKNOWN",
		Desc: "Unable to determine connection status.",
	},
}

// Statuses obtains possible statuses.
func Statuses() []StatusInfo {
	return statuses
}

// Connectivity exposes connection functionality.
type Connectivity interface {

	// Status obtains the current connection status, returning an error on failure.
	Status() (Status, error)

	// Reconnect reconnects, returning success/failure represented via a boolean.
	Reconnect() bool
}
