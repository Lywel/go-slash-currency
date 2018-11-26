package endpoint

const (
	inboundDir uint = iota
	outboundDir
	other
)

type message struct {
	Type     string      `json:"type"`
	Data     interface{} `json:"data"`
	DataType string      `json:"dataType"`
}
