package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/trains629/ms/base"
)

var (
	_name    = flag.String("name", "writer", "service name")
	_host    = flag.String("host", "127.0.0.1", "host")
	_port    = flag.Int64("port", 8051, "port")
	_version = flag.String("version", "v1", "version")
)

func publish(nsq *base.ServiceConfig, topic string, dict *Dict) error {
	n, err := base.NewQueueProducer(":4150") //nsq.GetAddr()
	if err != nil {
		return err
	}
	defer n.Stop()
	start := time.Now()
	err = dict.ReadDict(func(val interface{}, b bool) {
		if val == nil {
			return
		}
		if bb, err := json.Marshal(val); err == nil {
			if err = n.Publish(topic, bb); err != nil {
				log.Println(70, err, val)
			}
		}
	})
	end := time.Now().Sub(start)
	log.Println("结束时间", err, end)
	return err
}

func getToken(service *base.ServiceConfig) string {
	val, ok := service.GetInfoValue().(map[string]interface{})
	if !ok {
		return ""
	}
	topic, ok := val["topic"]
	return topic.(string)
}

func run(r *base.Runner) error {
	service := base.ReadServiceInfo(r.Ctx, r.Cli, "writer")
	if service == nil {
		return fmt.Errorf("error: %s", "writer nil")
	}
	nsq := base.ReadServiceInfo(r.Ctx, r.Cli, "nsq")
	if nsq == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}
	topic := getToken(service)
	dict := NewDict("../dict", DictEN)
	err := publish(nsq, topic, dict)
	log.Println(err)
	<-time.After(6 * time.Second)
	r.Cancel()
	return nil
}

func main() {
	runner, err := base.NewEtcdRunner()
	if err != nil {
		log.Fatalln(err)
	}
	if err := runner.Run(run); err != nil {
		log.Fatalln(err)
	}
}
