package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
)

const (
	//DictEN 英汉字典
	DictEN = iota
	// DictZH 汉英字典
	DictZH
)

//DictExecer 数据库sql执行接口
type DictExecer interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
}

//DictInserter 数据库插入接口
type DictInserter interface {
	Insert(tx DictExecer) error
}

type dictOffset struct {
	Word   string
	Offset int64
	count  int
	length int64
}

func (do *dictOffset) Count() int {
	if do.count <= 0 {
		return int(do.length - do.Offset)
	}

	return do.count
}

// Dict 字典对象
type Dict struct {
	path string
	kind uint
}

func (d *Dict) Path(ext string) string {
	kind := "en"
	if d.kind == DictZH {
		kind = "zh"
	}
	return path.Join(d.path, kind+"."+ext)
}

func (d *Dict) value(l *[]string) interface{} {
	switch d.kind {
	case DictEN:
		return en(l)
	case DictZH:
		return nil //zh(l)
	}
	return nil
}

func (d *Dict) size(ext string) int64 {
	f, err := os.Stat(d.Path(ext))
	if err != nil {
		return 0
	}
	return f.Size()
}

func (d *Dict) ReadInd(str string) ([]interface{}, error) {
	of, err := os.Open(d.Path("ind"))
	if err != nil {
		return nil, err
	}
	defer of.Close()
	scan := bufio.NewScanner(of)
	ll := []interface{}{}
	str = strings.ToLower(str)
	err = d.offsetKey(scan, str, d.size("z"), func(d1 *dictOffset) {
		t2, err := d.offsetWord(d1)
		if err != nil {
			return
		}
		ll = append(ll, t2)
	})
	if err != nil {
		return nil, err
	}

	return ll, nil
}

func findString(pattern string, s string) bool {
	// regexp.MatchString(".*"+pattern+".*", s)

	return strings.Index(s, pattern) >= 0 //pattern == s
}

func (d *Dict) offsetKey(scan *bufio.Scanner, str string, flen int64, fun func(*dictOffset)) error {
	if !scan.Scan() {
		return fmt.Errorf("not scan")
	}
	var (
		pWord string
		pNo   int64
		word  string
		no    int64
	)

	pWord, pNo = readWord(scan)
	count := 0
	for scan.Scan() {
		word, no = readWord(scan)
		if no == 0 {
			continue
		}
		if findString(str, pWord) {
			count++
			fun(&dictOffset{Word: pWord, Offset: pNo, count: int(no - pNo), length: flen})
		}
		pWord, pNo = word, no
	}
	if findString(str, pWord) {
		count++
		fun(&dictOffset{pWord, pNo, int(no - pNo), flen})
	}

	if count > 0 {
		return nil
	}
	return fmt.Errorf("no find")
}

func (dict *Dict) offsetWord(d1 *dictOffset) (interface{}, error) {
	p := dict.Path("z")
	zf, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer zf.Close()
	str := readData(zf, d1)

	if str == "" {
		return nil, fmt.Errorf("empty string")
	}
	ll := strings.Split(str, "|")
	ll[0] = d1.Word
	return dict.value(&ll), nil
}

func (dict *Dict) ReadDict(fun func(interface{}, bool)) error {
	of, err := os.Open(dict.Path("ind"))
	if err != nil {
		return err
	}
	defer of.Close()
	p := dict.Path("z")
	zf, err := os.Open(p)
	if err != nil {
		return err
	}
	defer zf.Close()
	err = readList(bufio.NewScanner(of), dict.size("z"), func(do *dictOffset, next bool) {
		str := readData(zf, do)

		if str == "" {
			return
		}
		ll := strings.Split(str, "|")
		ll[0] = do.Word
		fun(dict.value(&ll), next)
	})
	return err
}

func readData(f *os.File, do *dictOffset) string {
	buf := make([]byte, do.Count())
	if n, err := f.ReadAt(buf, do.Offset); err != nil || n < do.Count() {
		return ""
	}
	r, err := zlib.NewReader(bytes.NewReader(buf))
	if err != nil {
		return ""
	}
	defer r.Close()
	buf1 := &bytes.Buffer{}
	if _, err = io.Copy(buf1, r); err != nil {
		return ""
	}
	return buf1.String()
}

func readList(scan *bufio.Scanner, flen int64, fun func(*dictOffset, bool)) error {
	if !scan.Scan() {
		return fmt.Errorf("not scan")
	}
	var (
		pWord string
		pNo   int64
		word  string
		no    int64
	)

	pWord, pNo = readWord(scan)
	for scan.Scan() {
		word, no = readWord(scan)
		if no == 0 {
			continue
		}
		fun(&dictOffset{Word: pWord, Offset: pNo, count: int(no - pNo), length: flen}, true)
		pWord, pNo = word, no
	}
	fun(&dictOffset{pWord, pNo, int(no - pNo), flen}, false)
	return nil
}

func readWord(scan *bufio.Scanner) (string, int64) {
	str := scan.Text()
	ll := strings.Split(str, "|")
	word := ll[0]
	id, err := strconv.ParseInt(ll[1], 10, 64)
	if err != nil {
		return "", 0
	}
	return word, id
}

func NewDict(path string, kind uint) *Dict {
	return &Dict{
		path: path,
		kind: kind,
	}
}
