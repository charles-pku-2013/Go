package main

import (
	"fmt"
	"os"
	"os/signal"
)

func main() {
	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt) // 如果第没有第二个参数，catch 所有的 signals

	// Block until a signal is received.
	for {
		s := <-c
		fmt.Println("Got signal:", s)
	}
}
