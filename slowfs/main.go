// slowfs is a fuse file system that contains nw files and supports creating
// new ones (that disappear into the void), with a delay while creating in
// order to reproduce a Go issue involving EINTR errors on Darwin.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"

	"golang.org/x/net/context"

	"github.com/jacobsa/fuse"
)

// Repeatedly create and join sub-processes.
func runStuffRepeatedly(ctx context.Context) (err error) {
	for {
		// Check for cancellation.
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return

		default:
		}

		err = exec.Command("sleep", "0.1").Run()
		if err != nil {
			err = fmt.Errorf("Run: %v", err)
			return
		}
	}
}

func openFilesRepeatedly(
	ctx context.Context,
	dir string) (err error) {
	for {
		// Check for cancellation.
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return

		default:
		}

		// Attempt to create a file.
		//
		// TODO(jacobsa): Are all of the flags necessary?
		var f *os.File
		f, err = os.OpenFile(
			path.Join(dir, "foo"),
			os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_TRUNC,
			0600)

		if err != nil {
			err = fmt.Errorf("OpenFile: %v", err)
			return
		}

		// Clean up.
		err = f.Close()
		if err != nil {
			err = fmt.Errorf("Close: %v", err)
			return
		}
	}
}

func run() (err error) {
	// Create a mount point.
	mountPoint, err := ioutil.TempDir("", "slowfs")
	if err != nil {
		err = fmt.Errorf("TempDir: %v", err)
		return
	}

	defer os.Remove(mountPoint)

	// Create and mount the file system.
	mfs, err := mount(mountPoint)
	if err != nil {
		err = fmt.Errorf("mount: %v", err)
		return
	}

	// When we're done, attempt to unmount and join the file system.
	defer func() {
		if err := fuse.Unmount(mountPoint); err != nil {
			log.Printf("fuse.Unmount: %v", err)
			return
		}

		if err := mfs.Join(context.Background()); err != nil {
			log.Printf("Join: %v", err)
			return
		}
	}()

	log.Printf("Ready: %s", mountPoint)

	// Try to exit cleanly on Ctrl-C.
	_, cancel := context.WithCancel(context.Background())

	interrupted := make(chan os.Signal, 1)
	signal.Notify(interrupted, os.Interrupt)
	<-interrupted
	cancel()

	return
}

func main() {
	log.SetFlags(log.Lmicroseconds | log.Lshortfile)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			"Usage: %s [flags] mount_point\n",
			os.Args[0])

		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	err := run()
	if err != nil {
		log.Fatal(err)
	}
}
