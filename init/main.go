package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"time"

	"github.com/trains629/ms/base"
)

var (
	_timeout     = flag.Int64("ttl", int64(2*time.Second), "timeout")
	_Endpoints   = base.StringArray{}
	_serviceName = flag.String("service", "", "service name")
	_port        = flag.Int64("port", 0, "service port")
	_host        = flag.String("host", "", "host")
	_config      = flag.String("config", "", "service config")
	_prefix      = flag.String("prefix", base.ServicePrefix, "prefix")
)

func registerService(serviceName string) error {
	if serviceName == "" {
		return errors.New("service name is empty")
	}
	cli, err := base.NewEtcdClient([]string(_Endpoints), time.Duration(*_timeout))
	if err != nil {
		return err
	}
	defer cli.Close()

	config := base.NewServiceConfig(serviceName)
	config.Prefix = *_prefix
	config.Host = *_host
	config.Port = *_port
	config.Info = map[string]interface{}{}
	if *_config != "" {
		var info interface{}
		if err := json.Unmarshal([]byte(*_config), &info); err == nil {
			config.Info = info
		}
	}

	if err = base.RegisterService(context.Background(), cli, config, 0); err != nil {
		return err
	}
	return nil
}

func main() {
	flag.Var(&_Endpoints, "endpoints", "endpoint")
	flag.Parse()
	b := base.CheckFunc(context.Background(), 10, func() bool {
		_, err := base.NewEtcdClient([]string(_Endpoints), time.Duration(*_timeout))
		return err == nil
	})

	if !b {
		log.Fatalln("error Service")
	}
	sName := *_serviceName
	// 没有服务名称的时候只是负责检查etcd服务
	if sName == "" {
		return
	}

	if err := registerService(sName); err != nil {
		log.Fatalln(err)
	}
}
