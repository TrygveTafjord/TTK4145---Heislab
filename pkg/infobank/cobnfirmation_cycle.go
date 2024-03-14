package infobank

// import (
// 	"fmt"

// 	"project.com/pkg/elevator"
// 	"project.com/pkg/network"
// 	"project.com/pkg/timer"
// )

// func confirmNewAssignmentFSM(
// 	networkUpdateTx_ch chan network.Msg,
// 	networkUpdateRx_ch chan network.Msg,
// 	btn elevator.ButtonEvent,
// 	orderConfirmed_ch chan bool,
// 	elevatorMap map[string]elevator.Elevator,
// 	thisElevator elevator.Elevator,
// ) {

// 	timeOut_ch := make(chan bool)

// 	const CONFIRM_TIME float64 = 0.1
// 	confirmedNodes := make(map[string]elevator.Elevator)
// 	var isConfirmed bool = false

// 	thisElevator.Requests[btn.Floor][btn.Button] = true
// 	thisElevator.OrderCounter++

// 	msg := network.Msg{
// 		MsgType:  network.NewOrder,
// 		Elevator: thisElevator,
// 	}

// 	networkUpdateTx_ch <- msg

// 	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)

// 	for {
// 		select {

// 		case Msg := <-networkUpdateRx_ch:

// 			if Msg.MsgType != network.ConfirmedOrder {
// 				fmt.Printf("Recieved a non-confirming message!\n")
// 				break
// 			}

// 			confirmedNodes[Msg.Elevator.Id] = Msg.Elevator
// 			elevatorMap[Msg.Elevator.Id] = Msg.Elevator

// 			if len(confirmedNodes) == len(elevatorMap)-1{
// 				isConfirmed = true
// 			}

// 		case <- timeOut_ch:
// 			orderConfirmed_ch <- isConfirmed
// 			return

// 	}
// 	}
// }

// func confirmFSM(
// 		networkUpdateTx_ch chan network.Msg,
// 		networkUpdateRx_ch chan network.Msg,
// 		orderConfirmed_ch chan bool,
// 		elevatorMap map[string]elevator.Elevator,
// 		thisElevator elevator.Elevator,
// 		elevStatusUpdate_ch chan elevator.Elevator,
// 	) {

// 		timeOut_ch := make(chan bool)

// 		const CONFIRM_TIME float64 = 0.1
// 		confirmedNodes := make(map[string]elevator.Elevator)
// 		var isConfirmed bool = false

// 		msg := network.Msg{
// 			MsgType:  network.ConfirmedOrder,
// 			Elevator: thisElevator,
// 		}

// 		go timer.Run_timer(CONFIRM_TIME, timeOut_ch)

// 		for {
// 			select {

// 			case recievedMsg := <-networkUpdateRx_ch:
// 				if recievedMsg.MsgType == network.OrderCompleted {
// 					handleOrderCompleted(elevatorMap, &msg.Elevator, &thisElevator)
// 					elevStatusUpdate_ch <- thisElevator
// 				}

// 				if recievedMsg.MsgType != network.ConfirmedOrder && recievedMsg.MsgType != network.NewOrder {
// 					break
// 				}
// 				confirmedNodes[recievedMsg.Elevator.Id] = recievedMsg.Elevator
// 				elevatorMap[recievedMsg.Elevator.Id] = recievedMsg.Elevator

// 				if len(confirmedNodes) == len(elevatorMap)-1{
// 					isConfirmed = true
// 				}
// 				if recievedMsg.MsgType == network.NewOrder{
// 					networkUpdateTx_ch <- msg
// 				}

// 			case <- timeOut_ch:
// 				orderConfirmed_ch <- isConfirmed
// 				return
// 				}
// 			}
// 		}


