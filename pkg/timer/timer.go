package timer

import (
	"time"
)

var (
	timerEndTime float64
	timerActive  bool
)

func Get_wall_time() float64 {
	return float64(time.Now().UnixNano()) * 1e-9
}

func Run_timer(duration float64, timerFinished chan bool) {
	timerEndTime = Get_wall_time() + duration
	for Get_wall_time() < timerEndTime {

	}
	timerFinished <- true
}

func Timer_stop() {
	timerActive = false
}

func Timer_timedOut() bool {
	return (timerActive && (Get_wall_time() > timerEndTime))
}

