package light

import (
	log "github.com/sirupsen/logrus"

	"github.com/kidoman/embd"
)

// // Light is the structure for a single Light attached to your GPIO
type Light struct {
	Name  string          `json:"name"`
	Color string          `json:"color"`
	GPIO  int             `json:"gpio"`
	Desc  string          `json:"desc"`
	State int             `json:"state"`
	Dpin  embd.DigitalPin `json:"-"`
}

//LightOn : Turns a light on through GPIO
func (l *Light) On(id string) error {
	log.Println("Called Light.On")
	// 	if  c.config.HighRelay { // A high relay uses High GPIO for close state
	// 		state = embd.Low
	// 	}
	// } else {
	// 	if !c.config.HighRelay {
	// 		state = embd.Low
	// 	}
	log.Println("Setting GPIO Pin:", l.GPIO, "On")
	pin, err := embd.NewDigitalPin(l.GPIO)
	if err != nil {
		return err
	}
	if err := pin.SetDirection(embd.Out); err != nil {
		return err
	}
	l.State = embd.High
	return pin.Write(embd.High)
}

//LightOff : Turns a light off through GPIO
func (l *Light) Off(id string) error {

	log.Println("Setting GPIO Pin:", l.GPIO, "Off")
	pin, err := embd.NewDigitalPin(l.GPIO)
	if err != nil {
		return err
	}
	if err := pin.SetDirection(embd.Out); err != nil {
		return err
	}
	l.State = embd.Low
	return pin.Write(embd.Low)
}
