package elevator


const (
	N_FLOORS  = 4
	N_BUTTONS = 3
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
	EB_Stopped
)

type Elevator struct {
  	Floor     	int
	Dirn      	MotorDirection
	Requests  	[N_FLOORS][N_BUTTONS]uint8
	Behaviour 	ElevatorBehaviour
	Stop_time   float64
}
