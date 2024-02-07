package fsm



import (
	"OTP.com/Heis2e/pkg/elevio"
	"OTP.com/Heis2e/pkg/timer"
	"fmt"
)

func FSM(Button_ch chan elevio.ButtonEvent, Floor_sensor_ch chan int, Stop_button_ch chan bool, Obstruction_ch chan bool){
	for {
		select {
		case Buttonevent := <-Button_ch:
			
		case Newfloor := <-Floor_sensor_ch:
			
		case Stopbutton := <-Stop_button_ch:
			
		case Obstruction := <-Obstruction_ch:
			
		}
	}
}