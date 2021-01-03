package accessory

import (
	"github.com/justinkiang/hc/service"
)

type SecuritySystem struct {
	*Accessory
	SecuritySystem *service.SecuritySystem
}

func NewSecuritySystem(info Info, currentValue int) *SecuritySystem {
	acc := SecuritySystem{}
	acc.Accessory = New(info, TypeSecuritySystem)
	acc.SecuritySystem = service.NewSecuritySystem()

	acc.SecuritySystem.SecuritySystemCurrentState.SetValue(currentValue)

	acc.AddService(acc.SecuritySystem.Service)

	return &acc
}
