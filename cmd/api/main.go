package main

import (
	"context"
	"hideout/api"
	apiconfig "hideout/cmd/api/config"
	"log"
	"runtime/debug"
	"time"
)

func main() {
	ctx := context.Background()

	apiconfig.Init(ctx)
	log.Println("Configuration was successfully loaded")

	go func() {
		t := time.Tick(time.Minute)
		for {
			<-t
			debug.FreeOSMemory()
		}
	}()

	api.Serve()
}
