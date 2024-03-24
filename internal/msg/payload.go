package msg

import "encoding/gob"

func init() {
	gob.Register(EmptyPayload{})
}

type EmptyPayload struct{}
