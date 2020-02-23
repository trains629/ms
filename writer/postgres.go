package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/nsqio/go-nsq"
	"github.com/trains629/ms/base"
	"go.etcd.io/etcd/clientv3"
	//yaml8 "sigs.k8s.io/yaml"
)

type WriterHandler struct {
	db  *sql.DB
	cli *clientv3.Client
}

func (w *WriterHandler) HandleMessage(m *nsq.Message) error {
	if len(m.Body) == 0 {
		return nil
	}
	// 插数据的时候需要使用事物，从消息队列中攒够十条一组去发送，
	// 然后在增加超时context，超时的时候就将攒的数据都发送
	// 然后再启动多个当前的服务
	if en := NewENWord(m.Body); en != nil {
		en.Insert(w.db)
	}
	// Returning a non-nil error will automatically send a REQ command to NSQ to re-queue the message.
	return nil
}

// TableSQL 建表语句
const TableSQL = `CREATE TABLE IF NOT EXISTS public.dict_en(
  id serial PRIMARY KEY,
  word text NOT NULL,
  pronunciation text[] NOT NULL,
  paraphrase text[] NOT NULL,
  rank text,
  pattern text,
  sentence jsonb,
  createtime timestamp without time zone NOT NULL DEFAULT LOCALTIMESTAMP,
  updatetime timestamp without time zone NOT NULL DEFAULT now()
)`

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
func (w *WriterHandler) Close() {
	if w.db != nil {
		w.db.Close()
	}
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
func NewPostgresHandler(ctx context.Context, cli *clientv3.Client) (*WriterHandler, error) {
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

	wr := &WriterHandler{db, cli}
	if err := wr.createTable(); err != nil {
		return nil, err
	}

	// 检查数据表是否创建
	return wr, nil
}
