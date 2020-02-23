package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trains629/ms/base"
	"go.etcd.io/etcd/clientv3"
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
	_name    = flag.String("name", "writer", "service name")
	_host    = flag.String("host", "127.0.0.1", "host")
	_port    = flag.Int64("port", 8051, "port")
	_version = flag.String("version", "v1", "version")
)

type RunFunc func(context.Context, *clientv3.Client, context.CancelFunc) error

func Run(fun RunFunc) error {
	flag.Parse()
	cli, err := base.NewEtcdClient(getEndpoints(), time.Duration(*_timeout))
	if err != nil {
		log.Fatalf("err %v", err)
		return err
	}
	defer cli.Close()
	ctx, cancel := context.WithCancel(context.Background())
	log.Println(44444)
	fun(ctx, cli, cancel)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	select {
	case <-cli.Ctx().Done():
	case <-ctx.Done():
	case <-stop:
		log.Println("exit")
	}
	cancel()
	<-time.After(6 * time.Second)
	return nil
}

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
			return
		}
	})

	end := time.Now().Sub(start)
	log.Println("结束时间", err, end)
	return err
}

func getToken(service *base.ServiceConfig) string {
	info, ok := service.Info.(map[string]interface{})
	if !ok {
		return ""
	}
	value, ok := info["value"]
	if !ok || value == nil {
		return ""
	}
	val, ok := value.(map[string]interface{})
	if !ok {
		log.Println(83, val, ok)
		return ""
	}
	topic, ok := val["topic"]

	log.Println(82, topic, ok)
	return topic.(string)
}

func run(ctx context.Context, cli *clientv3.Client, cancel context.CancelFunc) error {
	/*
		不需要注册，只需要读取字典的数据，然后读取消息队列的数据
	*/
	service := base.ReadServiceInfo(ctx, cli, "writer")
	log.Println(633, service)
	if service == nil {
		return fmt.Errorf("error: %s", "writer nil")
	}
	nsq := base.ReadServiceInfo(ctx, cli, "nsq")
	if nsq == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}
	topic := getToken(service)
	dict := NewDict("../dict", DictEN)
	err := publish(nsq, topic, dict)
	log.Println(111, err)
	<-time.After(6 * time.Second)
	log.Println("after")
	cancel()
	return nil
}

func main() {
	if err := Run(run); err != nil {
		log.Fatalf("err %v", err)
	}
}
