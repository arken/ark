package display

import (
	"fmt"
	"sync"
	"time"
)

var spinner = []string{"|", "/", "-", "\\"}

// SpinnerWait displays the actual spinner
func SpinnerWait(done chan int, message string, wg *sync.WaitGroup) {
	ticker := time.Tick(time.Millisecond * 128)
	frameCounter := 0
	for {
		select {
		case _ = <-done:
			wg.Done()
			return
		default:
			<-ticker
			ind := frameCounter % len(spinner)
			fmt.Printf("\r[%v] "+message, spinner[ind])
			frameCounter++
		}
	}
}
