// Package hc provides implementation of an IP transport for HomeKit accessories.
//
//     import (
//         "github.com/justinkiang/hc"
//         "github.com/justinkiang/hc/accessory"
//     )
//
//     acc := accessory.NewSwitch(...)
//     config := hc.Config{Pin: "00102003"}
//     t, err := hc.NewIPTransport(config, acc.Accessory)
//     ...
//     t.Start()
package hc
