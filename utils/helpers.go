package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func HandleErr(err error) {
	if err != nil {
		Log.Error(err)
		panic(err)
	}
}

func HandlePanic(crash bool) {
	if r := recover(); r != nil {
		Log.Error("Panic: %v++\n", r)
		if crash {
			panic(r)
		}
	}
}

func RemoveDuplicates(arr []string) []string {
	uniqueMap := make(map[string]struct{})
	var result []string

	for _, val := range arr {
		if _, exists := uniqueMap[val]; !exists {
			uniqueMap[val] = struct{}{}
			result = append(result, val)
		}
	}

	return result
}

func HandleCtrlC(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	Log.Warn("Press Ctrl+C to exit")
	go func() {
		<-c
		fmt.Println()
		Log.Warn("Ctrl+C pressed in Terminal")
		signal.Stop(c)
		close(c)
		cancel()
	}()
}

func CloneMap(original map[string]string) map[string]string {
	// Create a new map with the same type as the original
	cloned := make(map[string]string, len(original))

	// Copy each key-value pair from the original map to the cloned map
	for key, value := range original {
		cloned[key] = value
	}

	return cloned
}
