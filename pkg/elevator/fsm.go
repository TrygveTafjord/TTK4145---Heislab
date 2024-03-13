package elevator

import (
	//"fmt"
	//"fmt"
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
	prevelevator := *elevator
	obstruction := GetObstruction()

	for {
		select {
		case newElev := <- fromInfobank_ch:
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
			diagnose := Selfdiagnose(elevator, &prevelevator, obstruction)
			switch diagnose{
				case Healthy:
					elevator.Obstructed = false
					elevator.OrderCounter++
					elevStatusUpdate_ch <- *elevator
				case Obstructed:
					elevator.Obstructed = true
					elevator.OrderCounter++
					elevStatusUpdate_ch <- *elevator
				case Problem:
					//Reboot

				case Unchanged:
					//Nothing
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
		if requestShouldStop(*e) {
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
		e.Lights[floor][BT_Cab] = e.Requests[floor][BT_Cab]
		for btn := 0; btn < N_BUTTONS; btn++ {
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





func Selfdiagnose(elevator *Elevator, prevElevator *Elevator, obstruction bool)Diagnose {
	hasRequests := Check_request(*elevator)

	if hasRequests && elevator.Behaviour == prevElevator.Behaviour {

		switch elevator.Behaviour {
		case EB_Idle:
			*prevElevator = *elevator
			return Problem

		case EB_DoorOpen:
			if elevator.Floor == prevElevator.Floor {
				elevator.Standstill+= 1
			}
		case EB_Moving:
			if elevator.Floor == prevElevator.Floor {
				elevator.Standstill+=1
				}
		}
		*prevElevator = *elevator



		if elevator.Standstill > 10 &&  obstruction{
			return Obstructed
		} else if elevator.Standstill == 20 && !obstruction {
			return Problem
		}			

	}else if obstruction{
		*prevElevator = *elevator
		return Unchanged

	}else{
		elevator.Standstill = 0
		*prevElevator = *elevator
	}
	return Healthy
}

func Check_request(elevator Elevator)bool{
	for i := 0; i < N_FLOORS; i++ {
		for j := 0; j < N_BUTTONS; j++ {
			if elevator.Requests[i][j] {
				return true
			}
		}
	}
	return false
}
