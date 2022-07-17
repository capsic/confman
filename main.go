package main

import "os"

const DEFAULTHOME = "/opt/capsic/confman"

func main() {
	// Set CONFMANHOME
	defaultHome := true
	if len(os.Args[1:]) > 0 {
		if os.Args[1] != "" {
			os.Setenv("CONFMANHOME", os.Args[1])
			defaultHome = false
		}
	}
	if defaultHome {
		os.Setenv("CONFMANHOME", DEFAULTHOME)
	}

	a := App{}
	a.Initialize()
	a.Run()
}
