package spii8080bitbang

import (
	"errors"
	"machine"
	"time"
)

type TransferMode int

const (
	WriteOnly = TransferMode(0)
	WriteRead = TransferMode(1)
	ReadOnly = TransferMode(2)
)

// SoftSPI implements drivers.SPI but transfers the data using 8 bit parallel. It is "bit-banged", i.e. written in software without taking advantage of special hardware available on certain platforms, but works for all platforms. the Init method will configure all pins with the right pin modes.
//
// See https://www.displaymodule.com/blogs/knowledge/display-module-interfaces-explained-8080-6800-spi-i2c-rgb
type SoftSPI struct {
	DataPins []machine.Pin // Parallel data pins, should have a length of 8
	WRPin    machine.Pin // Write strobe pulse pin
	RDPin    machine.Pin // Read strobe pulse pin.
	ReadAccessTimeNs int // Controls Read Access Time, in nanaseconds. If you're interfacing with a display or something else that wont read any data, you can ignore this.
	TransferMode TransferMode //TransferMode controls what the [SoftSPI.Transfer] method does. If you're interfacing with a display or something else that wont read any data, you can ignore this.
}

func (b SoftSPI) Init() error {
	b.WRPin.Configure(machine.PinConfig{
		Mode: machine.PinOutput,
	})

	b.RDPin.Configure(machine.PinConfig{
		Mode: machine.PinOutput,
	})

	b.WRPin.High()
	b.RDPin.High()

	if len(b.DataPins) != 8 {
		return errors.New("invalid datapin length, must be 8")
	}

	for _, dp := range b.DataPins {
		dp.Configure(machine.PinConfig{
			Mode: machine.PinOutput,
		})
	}
	return nil
}

//Tx first transfers the bytes w, then reading len(r) bytes into r. If r is nil, Tx will only transfer and vice versa. It is optimized for write-only operations such as interfacing with displays.
func (i SoftSPI) Tx(w, r []byte) error {
	for _, v := range w {
		i.WRPin.Low()
		for idx, dp := range i.DataPins {
			mask := byte(1) << idx //00000001 << idx
			dp.Set((v & mask) != 0)
		}
		i.WRPin.High()
	}

	if r == nil {
		return nil
	}

	for _, dp := range i.DataPins {
		dp.Configure(machine.PinConfig{
			Mode: machine.PinInput,
		})
	}

	for bIdx := range r {
		i.RDPin.Low()
		if i.ReadAccessTimeNs != 0 {
			time.Sleep(time.Duration(i.ReadAccessTimeNs) * time.Nanosecond)
		}

		wb := byte(0)
		for pIdx, dp := range i.DataPins {
			if !dp.Get() {
				continue
			}

			mask := byte(1) << byte(pIdx) //00000001 << idx
			wb = wb & mask
		}

		r[bIdx] = wb
		i.RDPin.High()
	}

	for _, dp := range i.DataPins {
		dp.Configure(machine.PinConfig{
			Mode: machine.PinOutput,
		})
	}
	return nil
}

//Transfer calls [SoftSPI.Tx] with a single byte. how exactly it does it is configured with [SoftSPI.TransferMode]
func (i SoftSPI) Transfer(b byte) (byte, error) {
	if i.TransferMode == WriteOnly {
		err := i.Tx([]byte{b}, nil)
		return 0, err
	}

	wrBuf := []byte{b}
	if i.TransferMode == ReadOnly {
		wrBuf = nil
	}

	rdBuf := make([]byte, 1)
	err := i.Tx(wrBuf, rdBuf)

	return rdBuf[0], err
}