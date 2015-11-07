// Copyright (c) Microsoft Open Technologies, Inc.
// All Rights Reserved.
// Licensed under the Apache License, Version 2.0.
// See License.txt in the project root for license information.
package common

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
)

// This step creates the actual virtual machine.
//
// Produces:
//   VMName string - The name of the VM
type StepCreateVM struct {
	VMName          string
	SwitchName      string
	RamSizeMB       uint
	DiskSize        uint
	Generation      uint
	Cpu             uint
	EnabeSecureBoot bool
}

func (s *StepCreateVM) Run(state multistep.StateBag) multistep.StepAction {
	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Creating virtual machine...")

	path := state.Get("packerTempDir").(string)

	// convert the MB to bytes
	ram := int64(s.RamSizeMB * 1024 * 1024)
	diskSize := int64(s.DiskSize * 1024 * 1024)

	switchName := s.SwitchName
	enabeSecureBoot := s.EnabeSecureBoot

	err := driver.CreateVirtualMachine(s.VMName, path, ram, diskSize, switchName, s.Generation)
	if err != nil {
		err := fmt.Errorf("Error creating virtual machine: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	err = driver.SetVirtualMachineCpu(s.VMName, s.Cpu)
	if err != nil {
		err := fmt.Errorf("Error creating setting virtual machine cpu: %s", err)
		state.Put("error", err)
		ui.Error(err.Error())
		return multistep.ActionHalt
	}

	if s.Generation == 2 {
		err = driver.SetSecureBoot(s.VMName, enabeSecureBoot)
		if err != nil {
			err := fmt.Errorf("Error setting secure boot: %s", err)
			state.Put("error", err)
			ui.Error(err.Error())
			return multistep.ActionHalt
		}
	}

	// Set the final name in the state bag so others can use it
	state.Put("vmName", s.VMName)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	if s.VMName == "" {
		return
	}

	driver := state.Get("driver").(Driver)
	ui := state.Get("ui").(packer.Ui)
	ui.Say("Unregistering and deleting virtual machine...")

	err := driver.DeleteVirtualMachine(s.VMName)
	if err != nil {
		ui.Error(fmt.Sprintf("Error deleting virtual machine: %s", err))
	}
}