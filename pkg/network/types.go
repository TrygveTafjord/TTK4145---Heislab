package network

import "project.com/pkg/elevator"


type NewRequest struct {
	Id			string
	Request     elevator.ButtonEvent
}

type StateUpdate struct{
	Id			string			
	State		elevator.State
}

type RequestCleared struct{
	Id				string
	ClearedRequests	[]elevator.ButtonEvent
}

type Obstructed struct{
	Id			 string
	Obstructed   bool
}

type Confirm struct{
	Id			string
	PassWrd   	string
}