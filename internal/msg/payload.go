package msg

import (
	"encoding/gob"
)

func init() {
	gob.Register(EmptyPayload{})

	registerRequest()
	registerResponse()
}

type EmptyPayload struct{}
