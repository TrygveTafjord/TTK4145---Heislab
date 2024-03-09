package network

import "project.com/pkg/elevator"


type MsgType int
const (
	NewOrder		MsgType = 0
	OrderCompleted 	MsgType = 1
	StateUpdate 	MsgType = 2
)

type Msg struct {
	MsgType  	MsgType
	Elevator 	elevator.Elevator
}