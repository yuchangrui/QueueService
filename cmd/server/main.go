package main

import "QueueService/pkg/network"

func main() {
	s := network.Server{}
	s.Init()
	s.Start()
}
