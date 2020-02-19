package main

import (
	"fmt"
	"log"
	"strings"

	"database/sql"

	"github.com/trains629/ms/base"
	//yaml8 "sigs.k8s.io/yaml"
)

type postConfig struct {
	*base.ServiceConfig
}

func connStr(v map[string]interface{}) string {
	ll := make([]string, 0)
	for name, va1 := range v {
		v1, k2 := va1.(string)
		if k2 {
			//v1 = strings.Trim(v1,)
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
