package main

import (
	"log"

	"github.com/trains629/ms/base"
)

func main() {
	runner, err := base.NewEtcdRunner()
	if err != nil {
		log.Fatalln(err)
	}
	if err := runner.Run(run); err != nil {
		log.Println(err)
	}
}
