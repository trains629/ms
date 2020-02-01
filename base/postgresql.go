package base

import (
	"database/sql"

	_ "github.com/lib/pq"
)

//OpenPostgresqlClient 打开postgresql数据库
func OpenPostgresqlClient(conninfo string) (*sql.DB, error) {
	//var  = `dbname=biger user=biger password='hao123456789' sslmode=disable host=git.trains629.com`
	db, err := sql.Open("postgres", conninfo)
	if err != nil {
		return nil, err
	}
	return db, err
}
