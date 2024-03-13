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
	State 				elevator.State
}

type ObstructedMsg struct {
	Id 			string
	Obstructed	bool
}

type StateMsg struct {
	Id		string
	State	elevator.State
}

type RequestClearedMsg struct {
	Direction 	elevator.ElevatorBehaviour
	Floor		int
}