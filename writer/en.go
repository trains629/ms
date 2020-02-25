package main

import (
	"database/sql"
	"encoding/json"

	"github.com/lib/pq"
)

type EN struct {
	Word          string        `json:"word"`
	ID            int           `json:"id"`
	Pronunciation []string      `json:"pronunciation"`
	Paraphrase    []string      `json:"paraphrase"`
	Rank          string        `json:"rank"`
	Pattern       string        `json:"pattern"`
	Sentence      []interface{} `json:"sentence"`
}

type sqlExec interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

func (en *EN) Insert(exec sqlExec) error {
	sql := `insert into public.dict_en (word,pronunciation,paraphrase,sentence) values ($1,$2,$3,$4);`

	sentence, err := json.Marshal(en.Sentence)
	if err != nil {
		sentence = []byte{}
	}
	if _, err := exec.Exec(sql, en.Word, pq.Array(en.Pronunciation), pq.Array(en.Paraphrase),
		sentence); err != nil {
		return err
	}
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

// NewENWord 新词条
func NewENWord(data []byte) *EN {
	result := &EN{}
	if err := json.Unmarshal(data, result); err != nil {
		return nil
	}
	return result
}
