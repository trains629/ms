package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/trains629/ms/base"
	"go.etcd.io/etcd/clientv3"
	//yaml8 "sigs.k8s.io/yaml"
)

func newWriterService(topic string) *base.ServiceConfig {
	return &base.ServiceConfig{
		Name:    *_name,
		Prefix:  base.ServicePrefix,
		Version: *_version,
		Host:    *_host,
		Port:    *_port,
		Info:    map[string]string{"topic": topic},
	}
}

func consumer(ctx context.Context, nsqConf *base.ServiceConfig, post Handler, topic string) error {
	if nsqConf == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}

	if post == nil {
		return fmt.Errorf("error: %s", "post nil")
	}
	consumer1, err := base.NewConsumer(topic, "writer")
	if err != nil {
		return err
	}

	consumer1.AddHandler(post)
	consumer1.ConnectToNSQLookupd(nsqConf.GetAddr())
	select {
	case <-consumer1.StopChan:
		post.Close()
	case <-ctx.Done():
		consumer1.Stop()
		post.Close()
	}
	return fmt.Errorf("error: %s", "nsq stop")
}

func initService(ctx context.Context, cli *clientv3.Client, topic string) error {
	ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
	var post Handler
	var nsq *base.ServiceConfig
	var err error
	go func() {
		defer cancel()
		nsq = base.ReadServiceInfo(ctx1, cli, "nsq")
		if nsq == nil {
			err = fmt.Errorf("error: %s", "nsq nil")
			return
		}
		post, err = NewPostgresHandler(ctx, cli)
		log.Println("结束")
	}()
	select {
	case <-ctx1.Done():
		if ctx1.Err() != context.Canceled {
			return ctx1.Err()
		}
	}
	if err != nil {
		return err
	}
	return consumer(ctx, nsq, post, topic)
}
