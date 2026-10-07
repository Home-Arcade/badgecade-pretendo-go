package main

import (
	"sync"

	"github.com/PretendoNetwork/nintendo-badge-arcade-secure/nex"
)

var wg sync.WaitGroup

func main() {
	wg.Add(1)

	go startFileStore()
	go nex.StartNEXServer()

	wg.Wait()
}
