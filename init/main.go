package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/trains629/ms/base"
	//yaml8 "sigs.k8s.io/yaml"
)

func getEndpoints() []string {
	result := os.Args[1:]
	if len(result) <= 0 {
		return []string{"localhost:2379", "localhost:22379", "localhost:32379"}
	}
	return result
}

var (
	_timeout = flag.Int64("ttl", int64(2*time.Second), "timeout")
)

func main() {
	flag.Parse()
	cli, err := base.NewEtcdClient(getEndpoints(), time.Duration(*_timeout))
	if err != nil {
		log.Fatalf("err %v", err)
	}
	defer cli.Close()
	err = loadConf("../config.yaml", "flex", func(config *base.ServiceConfig) {
		log.Println(config.Name, config.GetKey())
		base.RegisterService(context.TODO(), cli, config, 0)
	})

	if err != nil {
		log.Fatal(err)
	}
}
