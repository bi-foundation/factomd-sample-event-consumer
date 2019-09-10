package main

import (
	"fmt"
	"github.com/bi-foundation/factomd-sample-event-consumer/events"
	"os"
	"os/signal"
	"sync"
)

var (
	eventServer = events.NewDefaultReceiver()
)

func main() {
	fmt.Printf("Starting event receiver on port %v\n", eventServer.GetAddress())
	go eventServer.Start()

	fmt.Printf("Press Ctrl+C to end\n")
	WaitForCtrlC()
	fmt.Printf("\n")

}

func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	var signal_channel chan os.Signal
	signal_channel = make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}
