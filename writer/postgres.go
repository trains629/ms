package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"database/sql"

	"github.com/nsqio/go-nsq"
	"github.com/trains629/ms/base"
	"go.etcd.io/etcd/clientv3"
)

type Handler interface {
	nsq.Handler
	//Close() error
}

type WriterHandler struct {
	db     *sql.DB
	cli    *clientv3.Client
	count  int
	tx     *sql.Tx
	ctx    context.Context
	cancel context.CancelFunc
	mx     *sync.Mutex
}

func (w *WriterHandler) init() {
	log.Println("初始化")
	var err error
	w.tx, err = w.db.Begin()
	if err != nil {
		return
	}
	w.tx.Exec("")

	w.ctx, w.cancel = context.WithTimeout(context.Background(), 1*time.Second)
	go func() {
		select {
		case <-w.ctx.Done():
			// 超时或计数满10
			log.Println(48, w.ctx.Err())
			if w.ctx.Err() == context.DeadlineExceeded {
				log.Println("超时结束")
			}
			w.mx.Lock()
			if w.tx != nil {
				log.Println("提交", w.count)
				w.tx.Commit()
			}
			w.count = 0
			w.mx.Unlock()
		}
		log.Println("结束循环")
	}()
}

const WriterHandlerLen = 100

func (w *WriterHandler) start() {
	w.mx.Lock()
	if w.count <= 0 {
		w.init()
	}
	w.count++
	w.mx.Unlock()
}

func (w *WriterHandler) stop() {
	w.mx.Lock()
	if w.count > WriterHandlerLen {
		w.cancel()
	}
	w.mx.Unlock()
}

func (w *WriterHandler) insert(body []byte) {
	w.start()
	en := NewENWord(body)
	if w.tx == nil || en == nil {
		return
	}
	en.Insert(w.tx)
	w.stop()
}

func (w *WriterHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) > 0 {
		w.insert(m.Body)
	}
	return nil
}

func (w *WriterHandler) createTable() error {
	/**
	检查数据表是否存在，不存在就创建数据表
	*/
	_, err := w.db.Exec(TableSQL)
	if err != nil {
		log.Println(6969, err)
		return err
	}
	return nil
}

// Close 关闭连接
func (w *WriterHandler) Close() error {
	if w.db != nil {
		return w.db.Close()
	}
	return nil
}

type postConfig struct {
	*base.ServiceConfig
}

func connStr(v map[string]interface{}) string {
	ll := make([]string, 0)
	for name, va1 := range v {
		v1, k2 := va1.(string)
		if k2 {
			if strings.Index(v1, " ") >= 0 {
				v1 = "'" + v1 + "'"
			}
			ll = append(ll, fmt.Sprintf(`%s=%s`, name, v1))
		}
	}
	// `dbname=biger user=biger password='hao123456789' sslmode=disable host=git.trains629.com`
	return strings.Join(ll, " ")
}

func (pc *postConfig) getConn() string {
	log.Println(96, pc.Info)
	val, ok := pc.Info.(map[string]interface{})
	if !ok {
		return ""
	}
	value, ok := val["value"]
	if !ok {
		return ""
	}
	if val1, ok := value.(map[string]interface{}); ok {
		log.Println(113, 111)
		return connStr(val1)
	}
	return ""
}

func (pc *postConfig) Open() (*sql.DB, error) {
	conn := pc.getConn()
	db, err := base.OpenPostgresqlClient(conn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// NewPostgresHandler 创建postgresql对象
func NewPostgresHandler(ctx context.Context, cli *clientv3.Client) (Handler, error) {
	conf := base.ReadServiceInfo(ctx, cli, "postgresql")
	if conf == nil {
		return nil, fmt.Errorf("error: %s", "post nil")
	}
	post := &postConfig{conf}
	if post == nil {
		return nil, fmt.Errorf("error: %s", "post nil")
	}
	db, err := post.Open()
	if err != nil {
		return nil, err
	}
	return NewWriterHandler(db, cli)
}

// NewWriterHandler 创建回调
func NewWriterHandler(db *sql.DB, cli *clientv3.Client) (Handler, error) {
	wr := &WriterHandler{db: db, cli: cli, mx: &sync.Mutex{}}
	// 检查数据表是否创建
	if err := wr.createTable(); err != nil {
		return nil, err
	}
	return wr, nil
}
