package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type handleFunc func(resp http.ResponseWriter, req *http.Request)

// PostHandler hander对象
type PostHandler struct {
	patterns map[string]handleFunc
}

func (p *PostHandler) setHeader(fun1 handleFunc) handleFunc {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Set("Access-Control-Allow-Origin", "*")  // 允许的域
		resp.Header().Set("Access-Control-Allow-Headers", "*") // 运行的headers Content-Type,Access-Token
		resp.Header().Set("Access-Control-Allow-Methods", "*") // 运行的动作 PUT,POST,GET,DELETE,OPTIONS
		resp.Header().Set("Content-Type", "application/json; charset=utf-8")
		fun1(resp, req)
	}
}

func (p *PostHandler) mainRoot(res http.ResponseWriter, req *http.Request) {

	hh := map[string]interface{}{
		"hello": "world",
	}
	log.Println(req.URL)
	hh["url"] = req.URL
	buf, err := json.Marshal(hh)
	if err != nil {
		res.Write([]byte(`{"error":"` + err.Error() + `"}`))
	}
	res.Write(buf)
}

func (p *PostHandler) root(res http.ResponseWriter, req *http.Request) {
	path := req.URL.Path
	log.Println(43, path)
	buf := []byte(path)
	for pattern, handle := range p.patterns {
		if b, err := regexp.Match(pattern, buf); err == nil && b {
			handle(res, req)
			return
		}
	}
	p.mainRoot(res, req)
}

func (p *PostHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	p.setHeader(p.root)(res, req)
}

func (p *PostHandler) handle(pattern string, handler handleFunc) {
	if p.patterns == nil {
		p.patterns = make(map[string]handleFunc)
	}
	p.patterns[pattern] = handler
}

func (p *PostHandler) getPostData(res http.ResponseWriter, req *http.Request) interface{} {
	decode := json.NewDecoder(req.Body)
	var tmp interface{}
	if err := decode.Decode(&tmp); err != nil {
		log.Println(69, err)
	}
	return tmp
}

func (p *PostHandler) rootCreate(res http.ResponseWriter, req *http.Request) {
	data := p.getPostData(res, req)
	buf, err := json.Marshal(data)
	if err != nil {
		res.Write([]byte(`{"create":"` + req.URL.Path + `"}`))
		return
	}
	res.Write(buf)
}

func (p *PostHandler) rootUpdate(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(`{"read":"` + req.URL.Path + `"}`))
}

func (p *PostHandler) rootRead(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(`{"read":"` + req.URL.Path + `"}`))
}

func (p *PostHandler) rootList(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(`{"list":"` + req.URL.Path + `"}`))
}

// NewPostHandler 新的操作对象
func NewPostHandler() *PostHandler {
	r := &PostHandler{}
	r.handle("/create", r.rootCreate)
	r.handle("/update", r.rootUpdate)
	r.handle("/read", r.rootRead)
	r.handle("/list", r.rootList)
	return r
}
