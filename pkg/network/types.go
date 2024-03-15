package network

import "project.com/pkg/elevator"

type NewRequest struct {
	Id      string
	Request elevator.ButtonEvent
}

type StateUpdate struct {
	Id    string
	State elevator.State
}

type RequestCleared struct {
	Id              string
	ClearedRequests []elevator.ButtonEvent
}

type Obstructed struct {
	Id         string
	Obstructed bool
}

type Periodic struct {
	Id       string
	State    elevator.State
	Requests [elevator.N_FLOORS][elevator.N_BUTTONS]bool
}

type Confirm struct {
	Id      string
	PassWrd string
}
