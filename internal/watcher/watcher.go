package watcher

import (
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

func WatchConfig(filePath string, onChange func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	configDir := "."
	if stat, err := os.Stat(filePath); err == nil && !stat.IsDir() {
		fmt.Println(stat)
		configDir = filePath[:len(filePath)-len(stat.Name())]
	}
	fmt.Println(configDir)
	if err := watcher.Add(configDir); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&(fsnotify.Write|fsnotify.Create) != 0 && event.Name == filePath {
				fmt.Printf("[kloudmate] Config file updated: %s", event.Name)
				onChange()
			}
		case err := <-watcher.Errors:
			fmt.Printf("[kloudmate] Watcher error: %v", err)
		}
	}
}
