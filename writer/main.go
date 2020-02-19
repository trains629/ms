package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"database/sql"

	"github.com/nsqio/go-nsq"
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

func readServiceInfo(ctx context.Context, cli *clientv3.Client, name string) *base.ServiceConfig {
	conf := base.NewServiceConfig(name)
	ll, err := base.GetServiceList(ctx, cli, conf, 2) // 返回两条
	if err != nil || len(ll) <= 0 {
		return nil
	}

	val := ll[0]
	log.Println(val, val.Info)
	return val
}

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

/**
读取nsq 的配置信息
*/

type WriterHandler struct {
	db *sql.DB
}

func (w *WriterHandler) HandleMessage(m *nsq.Message) error {
	// 在这里将消息队列接收的数据发送到数据库
	fmt.Println(1555, m.Body)
	if len(m.Body) == 0 {
		// Returning nil will automatically send a FIN command to NSQ to mark the message as processed.
		return nil
	}

	//err := processMessage(m.Body)
	log.Println(string(m.Body))

	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}

func consumer(ctx context.Context, nsqConf *base.ServiceConfig, db *sql.DB, topic string) error {
	if nsqConf == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}
	consumer1, err := base.NewConsumer(topic, "writer")
	if err != nil {
		return err
	}

	consumer1.AddHandler(&WriterHandler{db})

	consumer1.ConnectToNSQLookupd(nsqConf.GetAddr())
	log.Println(8444)
	select {
	case <-consumer1.StopChan:
		log.Println("结束消息对象")
	case <-ctx.Done():
		consumer1.Stop()
	}
	log.Println(9999)
	return fmt.Errorf("error: %s", "nsq stop")
}

func initService(ctx context.Context, cli *clientv3.Client, topic string) error {
	// 需要读取两个服务的数据
	ctx1, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	conf := readServiceInfo(ctx, cli, "postgresql")
	if conf == nil {
		return fmt.Errorf("error: %s", "post nil")
	}
	nsq := readServiceInfo(ctx1, cli, "nsq")
	if nsq == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}
	log.Println(conf, nsq)
	post := &postConfig{conf}
	cancel()
	select {
	case b, c := <-ctx1.Done():
		log.Println("结束", b, c, ctx1.Err())
	}
	log.Println(114)
	db, err := post.Open()
	if err != nil {
		return err
	}
	consumer(ctx, nsq, db, topic)
	return nil
}

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
	go func() {
		// 还需要将自己注册到系统上
		service := newWriterService(topic) // reader向topic为主题的频道发消息
		err := base.RegisterService(ctx, cli, service, 10)
		if err != nil {
			log.Println(err)
			cancel()
		}
	}()

	select {
	case <-cli.Ctx().Done():
	case <-ctx.Done():
	}
	cancel()
}
