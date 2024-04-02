package util

import (
	"fmt"

	"github.com/google/gopacket"
)

func EndpointToString(net, transport gopacket.Endpoint) string {
	return fmt.Sprintf(
		"%s:%s",
		net.String(), transport.String(),
	)
}

func ReverseFlow(flow gopacket.Flow) gopacket.Flow {
	flow, _ = gopacket.FlowFromEndpoints(flow.Dst(), flow.Src())
	return flow
}
