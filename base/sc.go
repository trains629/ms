package base

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"go.etcd.io/etcd/clientv3"
)

// ServicePrefix 默认前缀
const ServicePrefix = "flex"

// ServiceConfig 服务信息
type ServiceConfig struct {
	Name    string      `json:"-"`
	Prefix  string      `json:"-"`
	Info    interface{} `json:"value"`
	Version string      `json:"-"`
	Host    string      `json:"host"`
	Port    int64       `json:"port"`
}

var (
	_strFunc = map[string]func(*ServiceConfig, string){
		"host": func(config *ServiceConfig, val string) {
			config.Host = val
		},
		"version": func(config *ServiceConfig, val string) {
			config.Version = val
		},
	}
	_int64Func = map[string]func(*ServiceConfig, int64){
		"port": func(config *ServiceConfig, val int64) {
			config.Port = val
		},
	}
)

// GetServiceName 取得服务名称
func (sc *ServiceConfig) GetServiceName() string {
	if sc.Name == "" {
		return ""
	}
	list := []string{}
	if sc.Prefix != "" {
		list = append(list, sc.Prefix)
	}

	list = append(list, sc.Name)
	if sc.Version == "" {
		sc.Version = "v1"
	}
	list = append(list, sc.Version)
	str := strings.Join(list, "/")
	if strings.Index(str, "/") == 0 {
		return str
	}

	return "/" + str
}

// GetKey 返回服务键名
func (sc *ServiceConfig) GetKey() string {
	if sc.Name == "" {
		return ""
	}

	if sc.Port <= 0 {
		return sc.GetServiceName() + "/" + sc.Host
	}
	return fmt.Sprintf("%s/%s:%d", sc.GetServiceName(), sc.Host, sc.Port)
}

// GetAddr 返回服务器地址
func (sc *ServiceConfig) GetAddr() string {
	if sc.Port <= 0 {
		return ""
	}
	return fmt.Sprintf("%s:%d", sc.Host, sc.Port)
}

// GetValueString 返回配置信息字符串
func (sc *ServiceConfig) GetValueString() string {
	b, err := json.Marshal(sc)
	if err != nil {
		return ""
	}
	return string(b)
}

func (sc *ServiceConfig) setStringVal(lk string, val string) {
	if f, ok := _strFunc[lk]; ok {
		f(sc, val)
	}
}

func (sc *ServiceConfig) setInt64Val(lk string, val int64) {
	if f, ok := _int64Func[lk]; ok {
		f(sc, val)
	}
}

// SetValueWithKey 按属性名更新
func (sc *ServiceConfig) SetValueWithKey(k string, v interface{}) {
	lk := strings.ToLower(k)
	switch val := v.(type) {
	case int:
		sc.setInt64Val(lk, int64(val))
	case int64:
		sc.setInt64Val(lk, val)
	case string:
		sc.setStringVal(lk, val)
	}
}

var _reg *regexp.Regexp

func init() {
	_reg, _ = regexp.Compile(`^(.*?)($|:(\d+)$)`)
}

func (sc *ServiceConfig) parseHost(host string) {
	sc.Host = host
	if _reg == nil {
		return
	}
	ll := _reg.FindStringSubmatch(host)
	l := len(ll)
	if l <= 0 {
		return
	}
	sc.Host = ll[1]
	if l <= 2 {
		return
	}
	if port, err := strconv.ParseInt(ll[3], 10, 64); err == nil {
		sc.Port = port
	}
}

// NewServiceConfig 新建服务配置信息
func NewServiceConfig(name string) *ServiceConfig {
	return &ServiceConfig{
		Name:    name,
		Prefix:  ServicePrefix,
		Version: "v1",
	}
}

// Key2ServiceConfig 由key生成服务信息对象
func Key2ServiceConfig(key string) *ServiceConfig {
	list := strings.Split(key, "/")
	result := &ServiceConfig{}
	start := len(list) - 1
	if start < 0 {
		goto end1
	}
	result.parseHost(list[start])
	start--
	if start < 0 {
		goto end1
	}
	if list[start][0] == 'v' {
		result.Version = list[start]
	} else {
		goto name1
	}
	start--
	if start < 0 {
		goto end1
	}
name1:
	result.Name = list[start]
	result.Prefix = strings.Join(list[0:start], "/")
end1:
	return result
}

// ReadServiceInfo 读取服务信息
func ReadServiceInfo(ctx context.Context, cli *clientv3.Client, name string) *ServiceConfig {
	conf := NewServiceConfig(name)
	ll, err := GetServiceList(ctx, cli, conf, 2) // 返回两条
	if err != nil || len(ll) <= 0 {
		return nil
	}

	val := ll[0]
	log.Println(val, val.Info)
	return val
}
