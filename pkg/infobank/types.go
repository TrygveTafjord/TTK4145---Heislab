package infobank

import (
	"project.com/pkg/elevator"
)

type ElevatorInfo struct {
	Id                  string
	OrderClearedCounter int
	OrderCounter        int
	Requests            [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	Lights              [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	State               elevator.State
}
