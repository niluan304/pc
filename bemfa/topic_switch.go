package bemfa

import (
	"fmt"
)

type Switch struct {
	on  func() error
	off func() error
}

func NewSwitch(on, off func() error) *Switch {
	return &Switch{
		on:  on,
		off: off,
	}
}

func (s *Switch) Handle(msg string) error {
	switch msg {
	case MsgOn:
		return s.on()
	case MsgOff:
		return s.off()
	default:
		return fmt.Errorf("unknown message: %s", msg)
	}
}
