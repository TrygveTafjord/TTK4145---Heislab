package network

import (
	"time"

	"project.com/pkg/elevator"
)

type Msg struct {
  Id        	string
  Counter     int
  NewInfo     bool
  Floor     	int
  Dirn      	elevator.MotorDirection
  Requests  	[elevator.N_FLOORS][elevator.N_BUTTONS]uint8
  Behaviour 	elevator.ElevatorBehaviour
}



const _pollRateImAlive = 500 * time.Millisecond