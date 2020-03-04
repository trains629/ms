package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/trains629/ms/base"
	"gopkg.in/yaml.v2"
	//yaml8 "sigs.k8s.io/yaml"
)

func walkService(name string, v interface{}, prefix string) *base.ServiceConfig {
	config := &base.ServiceConfig{Name: name, Prefix: prefix}

	switch service := v.(type) {
	case map[interface{}]interface{}:
		val := make(map[string]interface{})
		for k, v := range service {
			if key, ok := k.(string); ok {
				config.SetValueWithKey(key, v)
				val[key] = v
			}
		}
		config.Info = val
	case map[string]interface{}:
		config.Info = service
		config.SetValueWithKey("host", service["host"])
		config.SetValueWithKey("port", service["port"])
	default:
		return nil
	}
	return config
}

func getConfigValueByKey(key string) (r interface{}) {
	str := os.Getenv(key)
	log.Println(41, key, str)
	if str == "" {
		return nil
	}
	err := json.Unmarshal([]byte(str), &r)
	if err == nil {
		return
	}

	log.Println(49, str, err)

	host, port, err := net.SplitHostPort(str)
	if err != nil {
		log.Println(54, err)
		return str
	}

	val := map[string]interface{}{
		"host": host,
		"port": 0,
	}

	if pp, err := strconv.ParseInt(port, 10, 64); err == nil {
		val["port"] = pp
	}
	log.Println(66, val)
	return val
}

func loadEnv(prefix string, fun func(*base.ServiceConfig)) error {
	list := map[string]string{
		"postgresql": "Flex_POSTGRES",
		"nsq":        "FLEX_NSQ",
		"redis":      "FLEX_REDIS",
	}

	ii := 0

	for name, envName := range list {
		val := getConfigValueByKey(envName)
		if val == nil {
			continue
		}
		if url, ok := val.(string); ok {
			val = map[string]interface{}{
				"host": url,
				"port": 0,
			}
		}
		if config := walkService(name, val, prefix); config != nil {
			log.Println(89, config)
			fun(config)
		}
		ii++
	}

	if ii == 0 {
		return fmt.Errorf("error: %s", "empty env")
	}

	return nil
}

func loadConf(buf *[]byte, prefix string, fun func(*base.ServiceConfig)) error {
	m1 := make(map[string]map[string]interface{})
	err := yaml.Unmarshal(*buf, &m1)
	if err != nil {
		return err
	}
	services, ok := m1["services"]
	if !ok {
		return fmt.Errorf("services nil")
	}
	for name, v2 := range services {
		if config := walkService(name, v2, prefix); config != nil {
			fun(config)
		}
	}
	return nil
}
