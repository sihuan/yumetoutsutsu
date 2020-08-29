package main

import "time"

func main() {
	yume := NewYume()

	go yume.Start()

	for {
		time.Sleep(1 * time.Minute)
	}
}
