package timer


import (
	"time"
)

var(
	timerEndTime float64
	timerActive bool
)

//get_wall_time returns the current amount of seconds (with great precision, mind you) since a fixed point in history (1970) 
func get_wall_time() float64 {
	return float64(time.Now().UnixNano()) * 1e-9
}

func timer_start (duration float64){
	timerEndTime = get_wall_time() + duration
	timerActive = true
}

//timer_stop in case we need to stop the timer prematurely, i think. 
func timer_stop(){
	timerActive = false
}

func timer_timedOut() bool {
	return (timerActive && get_wall_time() > timerEndTime)
}
