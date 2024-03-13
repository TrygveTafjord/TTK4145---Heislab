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
	State    State
	Requests [N_FLOORS][N_BUTTONS]bool
	Lights   [N_FLOORS][N_BUTTONS]bool
}

type State struct {
	Floor      int
	Dirn       MotorDirection
	Behaviour  ElevatorBehaviour
	Obstructed bool
}

type Diagnose int

const (
	Healthy Diagnose = iota
	Obstructed
	Problem
	Unchanged
)

type OldElevator struct {
	Id                  string
	OrderClearedCounter int
	OrderCounter        int
	Floor               int
	Dirn                MotorDirection
	Requests            [N_FLOORS][N_BUTTONS]bool
	Lights              [N_FLOORS][N_BUTTONS]bool
	Behaviour           ElevatorBehaviour
	Standstill          int
	Obstructed          bool
}
