// THIS FILE IS AUTO-GENERATED
package service

import (
	"github.com/justinkiang/hc/characteristic"
)

const TypeSwitch = "49"

type Switch struct {
	*Service

	On *characteristic.On
}

func NewSwitch() *Switch {
	svc := Switch{}
	svc.Service = New(TypeSwitch)

	svc.On = characteristic.NewOn()
	svc.AddCharacteristic(svc.On.Characteristic)

	return &svc
}
