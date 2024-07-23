package main

import (
	"log"
	"os"

	"sberhl/config"
	"sberhl/internal/worker"
	"sberhl/pkg/logger"
)

func main() {
	c, err := config.NewConfig()
	if err != nil {
		log.Fatalf("new config: %s", err.Error())
	}

	loggerFile := os.Stdout
	if c.Logger.File != "stdout" {
		loggerFile, err = os.OpenFile(
			c.Logger.File,
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0664,
		)
		if err != nil {
			log.Fatalf("create logger file %s: %s", c.Logger.File, err.Error())
		}
		defer func(loggerFile *os.File) {
			err := loggerFile.Close()
			if err != nil {
				log.Fatalf("close logger file %s: %s", c.Logger.File, err.Error())
			}
		}(loggerFile)
	}

	l := logger.New(c.Logger.Level, loggerFile)

	w := worker.NewWorker(c.Worker.URL, c.Worker.Timeout, c.Worker.Threads, l)
	err = w.Run()
	if err != nil {
		log.Fatalf("new worker: %s", err.Error())
	}
}
