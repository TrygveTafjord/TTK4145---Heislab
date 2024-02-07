package elevio

type ElevatorState string

const (
	Idle    ElevatorState = "Idle"
	Moving  ElevatorState = "Moving"
	Stopped ElevatorState = "Stopped"
)
