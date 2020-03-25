package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"github.com/trains629/ms/base"
	"go.etcd.io/etcd/clientv3"
)

var (
	_name    = flag.String("name", "writer", "service name")
	_host    = flag.String("host", "", "address of this writer node, (default to the OS hostname)")
	_port    = flag.Int64("port", 8051, "port")
	_version = flag.String("version", "v1", "version")
	_channel = flag.String("channel", "writer", "channel")
)

func newWriterService(topic string) *base.ServiceConfig {

	host := *_host
	if host == "" {
		var err error
		if host, err = os.Hostname(); err != nil {
			return nil
		}
	}

	return &base.ServiceConfig{
		Name:    *_name,
		Prefix:  base.ServicePrefix,
		Version: *_version,
		Host:    host,
		Port:    *_port,
		Info:    map[string]string{"topic": topic},
	}
}

func consumer(ctx context.Context, nsqConf *base.ServiceConfig, post Handler, topic string) error {
	if nsqConf == nil {
		return errors.New("error: nsq nil")
	}

	if post == nil {
		return errors.New("error: post nil")
	}
	consumer1, err := base.NewConsumer(topic, *_channel)
	if err != nil {
		return err
	}

	consumer1.AddHandler(post)
	if err := consumer1.ConnectToNSQLookupd(nsqConf.GetAddr()); err != nil {
		log.Println(499, err)
		return err
	}
	select {
	case <-consumer1.StopChan:
	case <-ctx.Done():
		consumer1.Stop()
	}
	post.(io.Closer).Close()
	return nil
}

func initService(ctx context.Context, cli *clientv3.Client, topic string) error {
	var nsq *base.ServiceConfig
	if ok := base.CheckFunc(ctx, 3, func() bool {
		nsq = base.ReadServiceInfo(ctx, cli, "nsqlookupd")
		return nsq != nil
	}); !ok {
		return errors.New("do not connected nsq")
	}
	post, err := NewPostgresHandler(ctx, cli)
	if err != nil {
		return err
	}
	return consumer(ctx, nsq, post, topic)
}

func run(runner *base.Runner) error {
	topic := *_name
	// 应该先等消息队列可以使用，再执行后面的操作
	// 检查三次消息队列是否存在
	go func() {
		err := initService(runner.Ctx, runner.Cli, topic)
		if err != nil {
			log.Println(944, err)
			runner.Cancel()
		}
	}()
	// reader向topic为主题的频道发消息
	go func(service *base.ServiceConfig) {
		if service == nil {
			runner.Cancel()
			return
		}
		err := base.RegisterService(runner.Ctx, runner.Cli, service, 10)
		if err != nil {
			log.Println(102, err)
			runner.Cancel()
		}
	}(newWriterService(topic))
	return nil
}
