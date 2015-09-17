// Start an open(2) on a file in a directory specified by the user, then
// optionally deliver a signal to the process. If the process survives, print
// the result of opening.
//
// Usage:
//     interrupt_open [--signal 30] dir
//
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"syscall"
	"time"
)

var fSignal = flag.Int("signal", 0, "Signal number. Zero means no signal.")

func main() {
	flag.Parse()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)

	// Parse arguments.
	args := flag.Args()
	if len(args) != 1 {
		fmt.Fprintf(os.Stderr, "Usage:\n  %s [--signal 30] dir\n", os.Args[0])
		os.Exit(1)
	}

	dir := args[0]

	// Find the local PID.
	pid := os.Getpid()
	log.Println("My PID:", pid)

	// Start opening.
	go func() {
		p := path.Join(dir, "some_file")
		log.Printf("Opening %s...", p)

		_, err := os.Create(p)
		fmt.Fprintf(os.Stderr, "Create error: %v\n", err)

		if err == nil {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}()

	// Deliver a signal after a moment if appropriate.
	if *fSignal != 0 {
		time.Sleep(200 * time.Millisecond)
		log.Printf("Delivering signal %d...", *fSignal)

		err := syscall.Kill(pid, syscall.Signal(*fSignal))
		if err != nil {
			log.Fatalf("syscall.Kill: %v", err)
		}
	}

	// Block forever, waiting for os.Create to return in the goroutine created
	// above.
	<-make(chan struct{})
}
