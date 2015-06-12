package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"path"
)

var fDir = flag.String("dir", "", "Path to directory in which to create files.")

func runStuff() {
	for {
		err := exec.Command("sleep", "0.1").Run()
		if err != nil {
			log.Fatalf("Run: %v", err)
		}
	}
}

func main() {
	flag.Parse()
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	log.Println("My PID:", os.Getpid())

	// Start a bunch of workers running background processes.
	const numWorkers = 64
	for i := 0; i < numWorkers; i++ {
		go runStuff()
	}

	// Repeatedly open a file, truncating it.
	for i := 0; i < numWorkers; i++ {
		go func() {
			for {
				f, err := os.OpenFile(
					path.Join(*fDir, "foo"),
					os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC,
					0600)

				if err != nil {
					log.Fatalln(err)
				}

				f.Close()
			}
		}()
	}

	<-make(chan struct{})
}
