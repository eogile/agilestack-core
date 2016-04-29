package main

import (
	"log"
	"sync"

	"github.com/eogile/agilestack-core/registry"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ldate | log.Ltime)
}

func main() {
	log.Print("AT BEGINNING")
	subscriber := registry.NewNatsSubscriber("http://agilestack-nats.agilestacknet:4222")
	subscriber.InitShutdownHook()

	log.Print("before server listening")

	/*
	 * Preventing the program to exit
	 */
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	waitGroup.Wait()

}
