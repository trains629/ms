package main

import (
	"encoding/json"
	"strconv"
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

func en(ll *[]string) interface{} {
	r := EN{}
	r.Word = (*ll)[0]
	if id, err := strconv.Atoi((*ll)[1]); err == nil {
		r.ID = id
	}
	r.Pronunciation = make([]string, 2)
	if (*ll)[2] != "" {
		r.Pronunciation[0] = (*ll)[2] //"美"
	}
	if (*ll)[3] != "" {
		r.Pronunciation[1] = (*ll)[3] //"英"
	}
	if (*ll)[4] != "" {
		r.Pronunciation[0] = (*ll)[4]
		r.Pronunciation[1] = (*ll)[4]
	}
	json.Unmarshal([]byte((*ll)[5]), &r.Paraphrase)
	r.Rank = (*ll)[6]
	r.Pattern = (*ll)[7]
	json.Unmarshal([]byte((*ll)[8]), &r.Sentence)
	return r
}
