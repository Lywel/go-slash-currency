package endpoint

const (
	inboundDir uint = iota
	outboundDir
)

type ibftEvent struct {
	Direction uint   `json:"direction"`
	Type      string `json:"type"`
	Data      []byte `json:"data"`
}
