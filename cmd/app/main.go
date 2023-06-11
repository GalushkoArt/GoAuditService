package main

import (
	"github.com/galushkoart/go-audit-service/internal/app"
	"os"
	"os/signal"
)

func main() {
	exit := make(chan bool)
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, os.Interrupt, os.Kill)
		<-signals
		exit <- true
	}()
	finished := make(chan bool, 1)
	app.Run(exit, make(chan bool, 1), finished)
	if !<-finished {
		os.Exit(1)
	}
}
