package network

import "project.com/pkg/elevator"

type MsgType int

const (
	NewOrder       MsgType = 0
	OrderCompleted MsgType = 1
	StateUpdatea    MsgType = 2
	ConfirmedOrder MsgType = 3
	PeriodicMsg    MsgType = 4
	ObstructedMsg  MsgType = 5
)

type Msg struct {
	MsgType  MsgType
	Elevator elevator.Elevator
}



//Er på papiret ikke nødvendig å sende State og andre requests, men mylig det kan bli feil i systemet pga rekkefølgen ting skjer med channels
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
