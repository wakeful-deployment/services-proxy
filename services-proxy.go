package main

import (
	"fmt"
	"github.com/wakeful-deployment/services-proxy/ctemplate"
	"github.com/wakeful-deployment/services-proxy/haproxy"
	"os"
	"os/signal"
	"syscall"
)

func handleSignals() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals)

	for s := range signals {
		switch s {
		case os.Interrupt, os.Signal(syscall.SIGTERM):
			logger.Printf("%q: stop everything", s)
			haproxy.Kill()
			ctemplate.Kill()
			os.Exit(0)
		case os.Signal(syscall.SIGHUP):
			fmt.Println("SIGHUP!")
			haproxy.Restart()
		case os.Signal(syscall.SIGCHLD):
			// Ignore.
		default:
			logger.Printf("WTF %T %#v", s, s)
		}
	}
}

func main() {
	go haproxy.Start()
	go ctemplate.Start()
	go handleSignals()
	select {} // wait forever
}
