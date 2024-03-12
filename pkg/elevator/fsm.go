package elevator

import (
	"fmt"
	"time"

	"project.com/pkg/timer"
)

func FSM(fromInfobank_ch chan Elevator, toInfobank_ch chan Elevator, elevInitFSM_ch chan Elevator) {

	floorSensor_ch := make(chan int)
	stopButton_ch := make(chan bool)
	obstruction_ch := make(chan bool)
	timer_ch := make(chan bool)
	selfCheck_ch := make(chan bool)

	go PollFloorSensor(floorSensor_ch)
	go PollStopButton(stopButton_ch)
	go PollObstructionSwitch(obstruction_ch)
	go PeriodicCheck(selfCheck_ch)

	elevator := new(Elevator)
	*elevator = <-elevInitFSM_ch
	fmt.Printf("FSM was initted \n")
	prevelevator := *elevator
	obstruction := GetObstruction()

	for {
		select {
		case newElev := <- fromInfobank_ch:
			fmt.Print("Mottat newElev in FSM \n")
			fmt.Printf("I get into the block that makes shit happen and here the newElev has an OC of %v and an OCC of %v \n", newElev.OrderCounter, newElev.OrderClearedCounter) 
			fmt.Printf("while the local elevator has an OC of %v and an OCC of %v \n", elevator.OrderCounter, elevator.OrderClearedCounter)
			if newElev.OrderCounter > elevator.OrderCounter {
				elevator.Requests = newElev.Requests
				elevator.Lights = newElev.Lights
				elevator.OrderCounter = newElev.OrderCounter
				fsmNewAssignments(elevator, timer_ch)
				toInfobank_ch <- *elevator
			}
			elevator.Lights = newElev.Lights
			elevator.OrderClearedCounter = newElev.OrderClearedCounter
			setAllLights(elevator)

		case newFloor := <-floorSensor_ch:
			fsmOnFloorArrival(elevator, newFloor, timer_ch)
			toInfobank_ch <- *elevator

		case <-stopButton_ch:
			HandleStopButtonPressed(elevator)
		case <-timer_ch:
			HandleDeparture(elevator, timer_ch, obstruction)
			toInfobank_ch <- *elevator

		case obstr := <-obstruction_ch:
			if !obstr && elevator.Behaviour == EB_DoorOpen {
				go timer.Run_timer(3, timer_ch)
			}
			obstruction = obstr

		case <-selfCheck_ch:
			kill := Selfdiagnose(elevator, &prevelevator)
			if kill {
				fmt.Printf("\n Selfdestruct")
				//Logikk for å koble oss av nettet
			}
		}
	}
}

func HandleDeparture(e *Elevator, timer_ch chan bool, obstruction bool) {
	if obstruction && e.Behaviour == EB_DoorOpen {
		go timer.Run_timer(3, timer_ch)
	} else {
		e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)

		switch e.Behaviour {

		case EB_DoorOpen:
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			go timer.Run_timer(3, timer_ch)

		case EB_Moving:
			SetMotorDirection(e.Dirn)
			SetDoorOpenLamp(false)

		case EB_Idle:
			SetDoorOpenLamp(false)
		}
	}
}

func fsmOnFloorArrival(e *Elevator, newFloor int, timer_ch chan bool) {

	e.Floor = newFloor
	SetFloorIndicator(newFloor)
	fmt.Printf("In fsmOnFloorArrival i think my behaviour is: %v \n", e.Behaviour)
	setAllLights(e)
	switch e.Behaviour {
	case EB_Moving:
		fmt.Print("I know i am moving")
		if requestShouldStop(*e) {
			fmt.Print("It says we should stop!! \n ")
			SetMotorDirection(MD_Stop)
			e.Dirn = MD_Stop // Ole added march 12, needed for re-init
			SetDoorOpenLamp(true)
			requests_clearAtCurrentFloor(e)
			e.OrderClearedCounter++
			go timer.Run_timer(3, timer_ch)
			e.Behaviour = EB_DoorOpen
			setAllLights(e)
		}
	}
}

func fsmNewAssignments(e *Elevator, timer_ch chan bool) {
	// Her mistenker jeg det kommer bugs -Per 08.03
	if e.Behaviour == EB_DoorOpen {
		if requests_shouldClearImmediately(*e) {

			requests_clearAtCurrentFloor(e)

			e.OrderClearedCounter++

			go timer.Run_timer(3, timer_ch)
		}
		setAllLights(e)
		return
	}

	e.Dirn, e.Behaviour = GetDirectionAndBehaviour(e)
	switch e.Behaviour {

	case EB_DoorOpen:
		SetDoorOpenLamp(true)
		go timer.Run_timer(3, timer_ch)
		requests_clearAtCurrentFloor(e)
		e.OrderClearedCounter++

	case EB_Moving:
		SetMotorDirection(e.Dirn)
	}
	setAllLights(e)
}

func HandleStopButtonPressed(e *Elevator) {
	switch e.Floor {
	case -1:
		SetMotorDirection(0)
	default:
		SetMotorDirection(0)
		SetDoorOpenLamp(true)
	}
	e.Behaviour = EB_Stopped
}

func setAllLights(e *Elevator) {
	for floor := 0; floor < N_FLOORS; floor++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			//e.Lights[floor][btn] = e.Requests[floor][btn]
			SetButtonLamp(ButtonType(btn), floor, e.Lights[floor][btn])
		}
	}
}

func PeriodicCheck(selfCheck_ch chan bool) {
	for {
		time.Sleep(1000 * time.Millisecond)
		selfCheck_ch <- true
	}
}

func Selfdiagnose(elevator *Elevator, prevElevator *Elevator) bool {
	hasRequests := Check_request(*elevator)

	if hasRequests && elevator.Behaviour == prevElevator.Behaviour {

		switch elevator.Behaviour {
		case EB_Idle:
		case EB_DoorOpen:
			if elevator.Floor == prevElevator.Floor {
				elevator.similarity += 1
			}
		case EB_Moving:
			if elevator.Floor == prevElevator.Floor {
				elevator.similarity += 1
			}
		}

		if prevElevator.similarity == 15 {
			return true
		}
	} else {
		elevator.similarity = 0
	}
	*prevElevator = *elevator
	return false
}

func Check_request(elevator Elevator) bool {
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < N_BUTTONS; j++ {
			if elevator.Requests[i][j] {
				return true
			}
		}
	}
	return false
}
