package connectivity

type Status int

const (
	Connected = iota
	Disconnected
	Connecting
	Unknown
)

type StatusInfo struct {
	Code Status `json:"code"`
	Name string `json:"name"`
	Desc string `json:"description"`
}

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

func Statuses() []StatusInfo {
	return statuses
}

type Connectivity interface {
	Status() (Status, error)
	Reconnect() bool
}