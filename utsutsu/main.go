package main

import "time"

func main() {
	utsutsu := NewUtsutsu()

	go utsutsu.Start()

	for {
		time.Sleep(1 * time.Minute)
	}
}
