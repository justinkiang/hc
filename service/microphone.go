// THIS FILE IS AUTO-GENERATED
package service

import (
	"github.com/justinkiang/hc/characteristic"
)

const TypeMicrophone = "112"

type Microphone struct {
	*Service

	Volume *characteristic.Volume
	Mute   *characteristic.Mute
}

func NewMicrophone() *Microphone {
	svc := Microphone{}
	svc.Service = New(TypeMicrophone)

	svc.Volume = characteristic.NewVolume()
	svc.AddCharacteristic(svc.Volume.Characteristic)

	svc.Mute = characteristic.NewMute()
	svc.AddCharacteristic(svc.Mute.Characteristic)

	return &svc
}
