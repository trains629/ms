package base

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// ServiceReg 服务注册对象
type ServiceReg struct {
	cli    *clientv3.Client
	config *ServiceConfig
	ctx    context.Context

	leaseID clientv3.LeaseID
	ttl     int64
	stop    chan bool
}

func (sr *ServiceReg) keepAlive(leaseID clientv3.LeaseID) error {
	ctx, cancel := context.WithCancel(sr.ctx)
	resp, err := sr.cli.Lease.KeepAlive(ctx, leaseID)
	defer func() {
		cancel()
	}()
	if err != nil {
		return err
	}

	for {
		select {
		case _, ok := <-resp:
			if !ok {
				return fmt.Errorf("error: %v", "resp nil")
			}
		case stop := <-sr.stop:
			if stop {
				sr.cli.Lease.Revoke(sr.ctx, leaseID)
				return nil
			}
		case <-sr.cli.Ctx().Done():
			sr.Stop()
			return sr.cli.Ctx().Err()
		case <-sr.ctx.Done():
			sr.Stop()
			return sr.ctx.Err()
		}
	}
}

type timeoutFunc func(context.Context) (interface{}, error)

func funcWitchTimeout(ctx1 context.Context, fun1 timeoutFunc) (interface{}, error) {
	ctx, cancel := context.WithTimeout(ctx1, 2*time.Second)
	result, err := fun1(ctx)
	cancel()
	// 返回的err会有3种情况:
	// 1. Canceled 未超时手动取消
	// 2. DeadlineExceeded 超时，需要重新执行，或者返回错误
	// 3. 其他的错误，直接返回错误
	log.Println(65, err, result)
	defer log.Println(66, err, result)
	if err == nil {
		return result, err
	}
	if err == context.Canceled {
		return result, nil
	} else if err == context.DeadlineExceeded {
		// 超时,在这里增加控制，可以让函数重新执行
	}
	return nil, err
}

func (sr *ServiceReg) putGrant(key string, val string, ttl int64) (clientv3.LeaseID, error) {
	rr, err := funcWitchTimeout(sr.ctx, func(ctx context.Context) (interface{}, error) {
		return sr.cli.Lease.Grant(ctx, ttl)
	})

	if err != nil {
		sr.leaseID = 0
		return 0, err
	}
	lease := rr.(*clientv3.LeaseGrantResponse)
	sr.leaseID = lease.ID
	_, err = sr.put(key, val, clientv3.WithLease(lease.ID))
	return lease.ID, err
}

func (sr *ServiceReg) put(key, val string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	rr, err := funcWitchTimeout(sr.ctx, func(ctx context.Context) (interface{}, error) {
		return sr.cli.Put(ctx, key, val, opts...)
	})
	if err != nil {
		return nil, err
	}
	resp := rr.(*clientv3.PutResponse)
	return resp, err
}

// Start 启动服务
func (sr *ServiceReg) Start() error {
	log.Println(68, 68)
	key := sr.config.GetKey()
	if key == "" {
		return fmt.Errorf("error: %s", "key is empty")
	}
	ttl := sr.ttl
	val := sr.config.GetValueString()
	// 使用etcd都需要增加超时判断吗？

	var err error
	if ttl > 0 {
		id, err := sr.putGrant(key, val, ttl)
		if err != nil {
			return err
		}
		return sr.keepAlive(id)
	}
	log.Println(83838, val)
	_, err = sr.put(key, val)
	log.Println(85, err)
	return err
}

// Stop 结束服务
func (sr *ServiceReg) Stop() {
	if sr.stop == nil {
		sr.stop = make(chan bool)
	}
	sr.stop <- true
	close(sr.stop)
}

// NewEtcdClient 新建etcd对象
func NewEtcdClient(Endpoints []string, DialTimeout time.Duration) (*clientv3.Client, error) {
	if DialTimeout <= 0 {
		DialTimeout = 2 * time.Second
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   Endpoints,
		DialTimeout: DialTimeout,
	})
	return cli, err
}

// NewServiceReg 服务注册对象
func NewServiceReg(ctx context.Context, config *ServiceConfig,
	cli *clientv3.Client, ttl int64) *ServiceReg {
	return &ServiceReg{
		config: config,
		cli:    cli,
		ttl:    ttl,
		ctx:    ctx,
	}
}

// GetServiceList 得到服务列表
func GetServiceList(ctx context.Context, cli *clientv3.Client,
	config *ServiceConfig, limit int64) ([]*ServiceConfig, error) {
	key := config.GetServiceName()
	opts := []clientv3.OpOption{}
	opts = append(opts, clientv3.WithPrefix())
	//clientv3.WithSort()
	if limit > 0 {
		opts = append(opts, clientv3.WithLimit(limit))
	}

	rr, err := funcWitchTimeout(ctx, func(ctx context.Context) (interface{}, error) {
		return cli.Get(ctx, key+"/", opts...)
	})
	resp := rr.(*clientv3.GetResponse)
	if err != nil {
		return nil, err
	}
	list1 := make([]*ServiceConfig, 0)
	for _, ev := range resp.Kvs {
		conf := Key2ServiceConfig(string(ev.Key))
		json.Unmarshal(ev.Value, &conf.Info)
		list1 = append(list1, conf)
	}
	return list1, nil
}

// RegisterService 注册服务
func RegisterService(ctx context.Context, cli *clientv3.Client,
	config *ServiceConfig, ttl int64) error {
	sr := NewServiceReg(ctx, config, cli, ttl)
	if ttl <= 0 {
		return sr.Start()
	}
	go func() {
		select {
		case <-ctx.Done():
			sr.Stop()
		}
	}()
	return sr.Start()
}
