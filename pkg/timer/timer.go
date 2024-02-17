package timer

import (
	"time"
)

var (
	timerEndTime float64
	timerActive  bool
)

// get_wall_time returns the current amount of seconds (with great precision, mind you) since a fixed point in history (1970)
func Get_wall_time() float64 {
	return float64(time.Now().UnixNano()) * 1e-9
}

func Run_timer(duration float64, timerFinished chan bool) {
	timerEndTime = Get_wall_time() + duration
	for Get_wall_time() < timerEndTime {

	}
	timerFinished <- true

}

// timer_stop in case we need to stop the timer prematurely, i think.
func Timer_stop() {
	timerActive = false
}

func Timer_timedOut() bool {
	return (timerActive && (Get_wall_time() > timerEndTime))
}

