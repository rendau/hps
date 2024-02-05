package main

import (
	"github.com/rendau/hps/internal/app"
)

func main() {
	a := &app.App{}
	a.Init()
	a.Start()
	a.Listen()
	a.Stop()
	a.Exit()
}
