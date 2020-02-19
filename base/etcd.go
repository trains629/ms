package base

import (
	"context"
	"encoding/json"
	"fmt"
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
	ctx, cancel := context.WithCancel(context.Background())
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
				sr.cli.Lease.Revoke(context.Background(), leaseID)
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

func (sr *ServiceReg) putGrant(key string, val string, ttl int64) (clientv3.LeaseID, error) {
	lease, err := sr.cli.Lease.Grant(context.Background(), ttl)
	if err != nil {
		return 0, err
	}
	sr.leaseID = lease.ID
	_, err = sr.cli.Put(context.Background(), key, val,
		clientv3.WithLease(lease.ID))
	return lease.ID, err
}

// Start 启动服务
func (sr *ServiceReg) Start() error {
	key := sr.config.GetKey()
	if key == "" {
		return fmt.Errorf("error: %s", "key is empty")
	}
	ttl := sr.ttl
	val := sr.config.GetValueString()
	var err error
	if ttl > 0 {
		id, err := sr.putGrant(key, val, ttl)
		if err != nil {
			return err
		}
		return sr.keepAlive(id)
	}
	_, err = sr.cli.Put(context.Background(), key, val)
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
	resp, err := cli.Get(ctx, key+"/", opts...)
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
