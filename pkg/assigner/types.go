package assigner

import (
	"project.com/pkg/elevator"
)

type AssignerInput struct {
	Id					string
	Requests            [elevator.N_FLOORS][elevator.N_BUTTONS]bool
	State 				elevator.State
}