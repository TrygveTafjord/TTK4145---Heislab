package fsm



import (
	"OTP.com/Heis2e/pkg/elevio"
	"OTP.com/Heis2e/pkg/timer"
	"OTP.com/Heis2e/pkg/elevator"
	"fmt"
)



func FSM(Button_ch chan elevio.ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool){
	ElevatorPtr := new(elevator.Elevator)
	
	for {
		select {
		case Buttonevent := <-Button_ch:
			
		case Newfloor := <-Floor_sensor_ch:
			
		case Stopbutton := <-Stop_button_ch:
			HandleStopButtonPressed(ElevatorPtr)
			//set state stopped
		case Obstruction := <-Obstruction_ch:
			
		}
	}
}

func HandleStopButtonPressed(e *elevator.Elevator){
	//stop motor and consider opening door
	switch e.Floor {
		case -1:
			//stop motor
			elevio.SetMotorDirection(0)
		default: 
			elevio.SetMotorDirection(0)
			elevio.SetDoorOpenLamp(true)
	}
	//set state stopped
	e.Behaviour = elevator.EB_Stopped	
}