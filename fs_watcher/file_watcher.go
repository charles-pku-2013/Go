/*
go get -u golang.org/x/sys/...
go get -u github.com/fsnotify/fsnotify
*/
package main

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"log"
)

var (
	path = flag.String("path", "", "Name of file/dir to watch")
)

// main
func main() {
	flag.Parse()

	if *path == "" {
		log.Fatalf("You must provide a file to watch!")
	}

	// creates a new file watcher
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("ERROR", err)
	}
	defer watcher.Close()

	// wait for go func end
	done := make(chan bool)

	go func() {
		for {
			select {
			// watch for events
			case event := <-watcher.Events:
				// fmt.Printf("EVENT! %#v\n", event)
				fmt.Printf("%s %s\n", event.Name, event.Op.String())

			// watch for errors
			case err := <-watcher.Errors:
				fmt.Println("ERROR", err)
			}
		}
	}()

	// out of the box fsnotify can watch a single file, or a single directory
	if err := watcher.Add(*path); err != nil {
		fmt.Println("ERROR", err)
	}

	// keep running
	<-done
}
