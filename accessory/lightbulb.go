package accessory

import (
	"github.com/justinkiang/hc/service"
)

type Lightbulb struct {
	*Accessory
	Lightbulb *service.Lightbulb
}

// NewLightbulb returns an light bulb accessory which one light bulb service.
func NewLightbulb(info Info) *Lightbulb {
	acc := Lightbulb{}
	acc.Accessory = New(info, TypeLightbulb)
	acc.Lightbulb = service.NewLightbulb()

	acc.AddService(acc.Lightbulb.Service)

	return &acc
}
