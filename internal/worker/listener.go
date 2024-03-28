package worker

import (
	"context"
	"slices"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	"github.com/onee-only/netrat/internal/config"
	"github.com/pkg/errors"
)

type ListenOptions struct {
	Device, PcapFile string

	SnapLen     int32
	Promiscuous bool
	BPFFilter   string

	Timeout       time.Duration
	CaptureLayers []gopacket.LayerType
}

func (o *ListenOptions) Validate() (*ListenOptions, error) {
	if o == nil {
		o = &ListenOptions{}
	}

	if o.Device == "" && o.PcapFile == "" {
		return nil, errors.New("listener: device name and pcap file not specified")
	}

	if len(o.CaptureLayers) == 0 {
		return nil, errors.New("listener: capture layer not specified")
	}

	slices.Sort(o.CaptureLayers)
	o.CaptureLayers = slices.Compact(o.CaptureLayers)

	if o.SnapLen <= 0 {
		o.SnapLen = config.PacketSnapLen
	}

	return o, nil
}

type Listener struct {
	opts *ListenOptions
}

func newListener(opts *ListenOptions) (l *Listener, err error) {
	opts, err = opts.Validate()
	if err != nil {
		return nil, err
	}

	l = &Listener{opts: opts}

	return
}

func (l *Listener) Listen(ctx context.Context) (_ <-chan gopacket.Layer, err error) {
	packetStream := make(chan gopacket.Layer, config.PacketStreamBufSize)

	var handle *pcap.Handle
	if l.opts.Device != "" {
		handle, err = pcap.OpenLive(l.opts.Device, l.opts.SnapLen, l.opts.Promiscuous, pcap.BlockForever)
		if err != nil {
			return nil, errors.Wrap(err, "listener: creating handle from device")
		}
	} else {
		handle, err = pcap.OpenOffline(l.opts.PcapFile)
		if err != nil {
			return nil, errors.Wrap(err, "listener: creating handle from pcap file")
		}
	}

	if l.opts.BPFFilter != "" {
		if err := handle.SetBPFFilter(l.opts.BPFFilter); err != nil {
			return nil, errors.Wrap(err, "listener: setting BPF filter")
		}
	}

	var timeout <-chan time.Time
	if l.opts.Timeout > 0 {
		timeout = time.NewTimer(l.opts.Timeout).C
	}

	packets := gopacket.NewPacketSource(handle, handle.LinkType()).Packets()

	go func() {
		defer close(packetStream)
		defer handle.Close()

		var packet gopacket.Packet
		for {
			select {
			case <-ctx.Done():
				return
			case <-timeout:
				return
			case packet = <-packets:
				if packet == nil {
					return
				}
			}

			for _, capLayer := range l.opts.CaptureLayers {
				if layer := packet.Layer(capLayer); layer != nil {
					packetStream <- layer
				}
			}
		}
	}()

	return packetStream, nil
}
