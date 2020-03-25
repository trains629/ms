package base

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.etcd.io/etcd/clientv3"
)

// ArrayString 字符串切片
type ArrayString []string

// Runner 启动对象
type Runner struct {
	Ctx    context.Context
	Cli    *clientv3.Client
	Cancel context.CancelFunc
}

// RunEtcdFunc 启动对象回调函数
type RunEtcdFunc func(*Runner) error

// Run 启动函数
func (r *Runner) Run(fun RunEtcdFunc) error {
	defer r.Cli.Close()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	if err := fun(r); err != nil {
		r.Cancel()
	}
	select {
	case <-r.Cli.Ctx().Done():
	case <-r.Ctx.Done():
	case <-stop:
	}
	r.Cancel()
	<-time.After(10 * time.Second) // 延迟十秒退出
	return nil
}

var (
	_timeout   *int64 //= flag.Int64("ttl", int64(2*time.Second), "timeout")
	_Endpoints = StringArray{}
)

func checkETCDService() bool {
	ii := 0
	return CheckFunc(context.TODO(), 3, func() bool {
		ii++
		cli, err := NewEtcdClient([]string(_Endpoints), time.Duration(*_timeout))
		defer cli.Close()
		// 通过使用get去读取数据，如果无法读取或者读取超时就返回false，否则就返回true
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		_, err = cli.Get(ctx, "/"+ServicePrefix)
		cancel()
		if err == nil {
			return true
		}
		if err == context.Canceled {
			return true
		}
		return false
	})
}

// NewEtcdRunner 创建运行器
func NewEtcdRunner() (*Runner, error) {
	_timeout = flag.Int64("ttl", int64(2*time.Second), "timeout")
	flag.Var(&_Endpoints, "endpoints", "endpoint")
	flag.Parse()
	cli, err := NewEtcdClient([]string(_Endpoints), time.Duration(*_timeout))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		Cli:    cli,
		Ctx:    ctx,
		Cancel: cancel,
	}, nil
}

// CheckFunc 按周期检查
func CheckFunc(ctx context.Context, num int, fun func() bool) bool {
	stop := make(chan bool)
	tl := 2 * time.Second // 两秒检测一次
	ii := 0
	timer1 := time.AfterFunc(tl, func() {
		ii++
		if ii > num {
			close(stop)
			return
		}
		stop <- fun()
	})

	for {
		select {
		case <-ctx.Done():
			return false
		case cc, ok := <-stop:
			log.Println(cc, ok)
			if !ok {
				return ok
			}
			if cc {
				return cc
			}
			timer1.Reset(tl)
		}
	}
}
