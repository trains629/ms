package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"

	"github.com/trains629/ms/base"
)

var (
	_config = flag.String("config", "", "config")
)

func run(runner *base.Runner) error {
	const _prefix = base.ServicePrefix
	item := func(config *base.ServiceConfig) {
		log.Println(config.Name)
		base.RegisterService(context.TODO(), runner.Cli, config, 0)
	}

	b1, err := ioutil.ReadFile(*_config)
	if err == nil {
		err = loadConf(&b1, _prefix, item)
	}
	if err != nil {
		err = loadEnv(_prefix, item)
	}

	if err != nil {
		runner.Cancel()
	}

	return err
}

func main() {
	runner, err := base.NewEtcdRunner()
	if err != nil {
		log.Fatalln(err)
	}

	if err = runner.Run(run); err != nil {
		log.Fatalln(err)
	}
}
