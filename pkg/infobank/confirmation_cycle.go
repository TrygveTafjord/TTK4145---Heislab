package infobank

import (
	"project.com/pkg/elevator"
	"project.com/pkg/network"
)

func confirmationCycleFSM(
	networkUpdateTx_ch chan network.Msg,
	networkUpdateRx_ch chan network.Msg,
	btn elevator.ButtonEvent,
	orderConfirmed_ch chan bool,
	elevatorMap map[string]elevator.Elevator,
	thisElevator elevator.Elevator,
) {
	thisElevator.Requests[btn.Floor][btn.Button] = true
	thisElevator.OrderCounter++

	msg := network.Msg{
		MsgType:  network.NewOrder,
		Elevator: thisElevator,
	}

	var confirmedNodes map[string]elevator.Elevator

	for {

		select {
		case Msg := <-networkUpdateRx_ch:
			if Msg.msgType == network.ConfirmedOrder {
				confirmedNodes[Msg.Elevator.Id] = Msg.Elevator

			}
		}
	}

}
