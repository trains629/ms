package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/trains629/ms/base"
)

func getEndpoints() []string {
	result := flag.Args() //os.Args[1:]
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

func main() {
	flag.Parse()
	cli, err := base.NewEtcdClient(getEndpoints(), time.Duration(*_timeout))
	if err != nil {
		log.Fatalf("err %v", err)
	}
	defer cli.Close()
	topic := *_name
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err = initService(ctx, cli, topic)
		if err != nil {
			log.Println(139, err)
			cancel()
		}
	}()
	// reader向topic为主题的频道发消息
	go func(service *base.ServiceConfig) {
		err := base.RegisterService(ctx, cli, service, 10)
		if err != nil {
			log.Println(149, err)
			cancel()
		}
	}(newWriterService(topic))
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
}
