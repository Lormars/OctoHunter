package watcher

import (
	"fmt"
	"log"

	"github.com/fsnotify/fsnotify"
)

func WatchFile(fileName string, doScan chan bool, callBack func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		fmt.Println("Watching file: ", fileName)
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
					callBack()
					<-doScan
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(fileName)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
