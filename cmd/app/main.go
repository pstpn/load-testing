package main

import (
	"log"
	"sberhl/internal/worker"

	"sberhl/config"
)

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	w := worker.NewWorker(c.Worker.URL, c.Worker.Timeout, c.Worker.Threads)

	err = w.Run()
	if err != nil {
		log.Fatal(err)
	}
}
