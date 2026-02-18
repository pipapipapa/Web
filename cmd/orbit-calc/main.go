package main

import (
	"log"

	"orbit-calc/internal/api"
)

func main() {
	log.Println("Starting Orbit Calculation System...")

	api.StartServer()

	log.Println("Server terminated")
}