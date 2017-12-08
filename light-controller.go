package main

import (
	"fmt"
	"log"

	"github.com/kidoman/embd"
)

const LightsBucket = "lights"

func (c *Controller) GetLight(name string) (Light, error) {
	for k, v := range c.config.Lights {
		if v.Name == name {
			// Found!
			return c.config.Lights[k], nil
		}
	}
	return nil, fmt.Errorf("No light named '%s' present", name)
}

//GetLights - gets all the lights and returns JSON
func (c *Controller) GetLights() ([]Light, error) {

	return c.config.Lights, nil
}

func (c *Controller) ConfigureLight(id string, on bool, value int) error {

	l, ok := c.config.Lights[id]
	if !ok {
		return fmt.Errorf("Light named: '%s' does noy exist", id)
	}

	if c.config.DevMode {
		log.Println("Dev mode on. Skipping:", id, "On:", on, "Value:", value)
		return nil
	}

	return c.doSwitching(l.GPIO, on)

}

//LightOn : Turns a light on through GPIO
func (c *Controller) LightOn(id string) error {

	log.Printf("Called controller.light.LightOn for %v", id)
	l, ok := c.config.Lights[id]
	if !ok {
		return fmt.Errorf("Light named: '%s' does not exist", id)
	}

	l.State = "on"
	return c.doSwitching(l.GPIO, true)

}

//LightOff : Turns a light off through GPIO
func (c *Controller) LightOff(id string) error {

	l, ok := c.config.Lights[id]
	if !ok {
		return fmt.Errorf("Light named: '%s' does not exist", id)
	}

	l.State = "off"
	return c.doSwitching(l.GPIO, false)

}

func (c *Controller) CreateLight(light Light) error {

	log.Println("Entering AddLight")
	fn := func(id string) interface{} {
		light.ID = id
		tempDPin, err := embd.NewDigitalPin(light.GPIO)
		if err != nil {
			log.Printf("light.go: creating new dpin bombed\n")
			panic(err)
		}
		light.Dpin = tempDPin

		if err := light.Dpin.SetDirection(embd.Out); err != nil {
			log.Printf("light.go: light.dpin.SetDirection(embd.Out) failed - just a warning\n")
		}
		return light
	}
	return c.store.Create(LightsBucket, fn)
}
