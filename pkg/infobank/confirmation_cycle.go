package infobank

import (
	"fmt"
	"time"

	"project.com/pkg/elevator"
	"project.com/pkg/network"
	"project.com/pkg/timer"
)

func confirmationCycleFSM(
	networkUpdateTx_ch chan network.Msg,
	networkUpdateRx_ch chan network.Msg,
	btn elevator.ButtonEvent,
	orderConfirmed_ch chan bool,
	elevatorMap map[string]elevator.Elevator,
	thisElevator elevator.Elevator,
) {

	timeOut_ch := make(chan bool)
	stopSending_ch := make(chan bool)

	
	const CONFIRM_TIME float64 = 0.2
	confirmedNodes := make(map[string]elevator.Elevator)
	var isConfirmed bool = false

	thisElevator.Requests[btn.Floor][btn.Button] = true
	thisElevator.OrderCounter++

	msg := network.Msg{
		MsgType:  network.NewOrder,
		Elevator: thisElevator,
	}

	go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
	go sendPeriodicly(networkUpdateTx_ch, msg, stopSending_ch, CONFIRM_TIME)
	

	for {
		select {

		case Msg := <-networkUpdateRx_ch:
			fmt.Printf("motat melding \n")

			if Msg.MsgType != network.ConfirmedOrder {
				break
			}
			confirmedNodes[Msg.Elevator.Id] = Msg.Elevator
			elevatorMap[Msg.Elevator.Id] = Msg.Elevator

			if len(confirmedNodes) == len(elevatorMap)-1{
				fmt.Printf("bekrefter knappetrykk \n")
				isConfirmed = true
			}

		case <- timeOut_ch:
			fmt.Printf("time out \n")
			stopSending_ch <- true
			orderConfirmed_ch <- isConfirmed
			return
			}
		}
	}


func confirmFSM(
		networkUpdateTx_ch chan network.Msg,
		networkUpdateRx_ch chan network.Msg,
		orderConfirmed_ch chan bool,
		elevatorMap map[string]elevator.Elevator,
		thisElevator elevator.Elevator,
	) {
	
		timeOut_ch := make(chan bool)
		stopSending_ch := make(chan bool)


		const CONFIRM_TIME float64 = 0.2
		confirmedNodes := make(map[string]elevator.Elevator)
		var isConfirmed bool = false
	
		msg := network.Msg{
			MsgType:  network.ConfirmedOrder,
			Elevator: thisElevator,
		}
	
		go timer.Run_timer(CONFIRM_TIME, timeOut_ch)
		go sendPeriodicly(networkUpdateTx_ch, msg, stopSending_ch, CONFIRM_TIME)
		
		for {
			select {
	
			case Msg := <-networkUpdateRx_ch:
				fmt.Printf("motat melding \n")
				
				if Msg.MsgType != network.ConfirmedOrder && Msg.MsgType != network.NewOrder {
					break
				}				
				confirmedNodes[Msg.Elevator.Id] = Msg.Elevator
				elevatorMap[Msg.Elevator.Id] = Msg.Elevator

				if len(confirmedNodes) == len(elevatorMap)-1{
					fmt.Printf("bekrefter knappetrykk \n")
					isConfirmed = true
				}

			case <- timeOut_ch:
				fmt.Printf("timer out \n")
				stopSending_ch <- true
				orderConfirmed_ch <- isConfirmed
				return
				}
			}
		}


	func sendPeriodicly(networkUpdateTx_ch chan network.Msg, msg network.Msg, timeOut_ch chan bool, CONFIRM_TIME float64){
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		
		for {
			select {
			case <- timeOut_ch:
				fmt.Printf("ferdig å sende \n")
				return

			case <-ticker.C:
				fmt.Printf("sender Melding \n")
				networkUpdateTx_ch <- msg
			}
		}
	}


