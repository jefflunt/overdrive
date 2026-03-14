package main

import (
	"log"
	"os"

	"overdrive/api"
)

func main() {
	if err := api.RecordSchedulerStartTime(); err != nil {
		log.Printf("Failed to record scheduler start time: %v", err)
	}
	// Ensure projects directory exists
	if err := os.MkdirAll("projects", 0755); err != nil {
		log.Fatalf("Failed to create projects directory: %v", err)
	}

	// Cleanup old working jobs
	api.CleanupZombieJobs()

	// Start the scheduler
	log.Println("Starting scheduler...")
	api.StartScheduler()

	// Keep the process running
	select {}
}
