package gossipnet

import (
	"encoding/json"
)

// Network message structure
type Message struct {
	Data []byte
	Type int
}

func (msg *Message) Encode() ([]byte, error) {
	return json.Marshal(msg)
}

func (msg *Message) Decode(data []byte) error {
	return json.Unmarshal(data, msg)
}
