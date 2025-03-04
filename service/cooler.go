package service

import (
	"github.com/justinkiang/hc/characteristic"
)

type Cooler struct {
	*HeaterCooler

	CoolingThresholdTemperature *characteristic.CoolingThresholdTemperature
}

func NewCooler() *Cooler {
	svc := Cooler{}

	svc.CoolingThresholdTemperature = characteristic.NewCoolingThresholdTemperature()
	svc.AddCharacteristic(svc.CoolingThresholdTemperature.Characteristic)

	return &svc
}
