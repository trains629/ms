package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/trains629/ms/base"
)

var (
	_name    = flag.String("name", "writer", "service name")
	_host    = flag.String("host", "127.0.0.1", "host")
	_port    = flag.Int64("port", 8051, "port")
	_version = flag.String("version", "v1", "version")
)

func get1(endpoint string, v interface{}) error {
retry:
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/vnd.nsq; version=1.0")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	log.Println(string(body))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 && !strings.HasPrefix(endpoint, "https") {
			//endpoint, err = httpsEndpoint(endpoint, body)
			if err != nil {
				return err
			}
			goto retry
		}
		return fmt.Errorf("got response %s %q", resp.Status, body)
	}
	err = json.Unmarshal(body, &v)

	if err != nil {
		return err
	}

	return nil
}

func publish(nsq *base.ServiceConfig, topic string, dict *Dict) error {
	log.Println(nsq.GetAddr())
	// 这个对象只支持访问nsq,如果要使用nsqlookupd，就需要增加函数去处理
	// 为什么消费者就做了使用nsqlookupd的函数，而生产者没有，是因为消费者太多的原因吗？
	n, err := base.NewQueueProducer(nsq.GetAddr())
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
				log.Println(70, err)
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
	nsq := base.ReadServiceInfo(r.Ctx, r.Cli, "nsqd")
	if nsq == nil {
		return fmt.Errorf("error: %s", "nsq nil")
	}
	var a1 interface{}

	err1 := get1("http://nsqlookupd-service/nodes", &a1)

	log.Println(err1, a1)

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
