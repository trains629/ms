package main

import (
	"fmt"
	"io/ioutil"

	"github.com/trains629/ms/base"
	"gopkg.in/yaml.v2"
	//yaml8 "sigs.k8s.io/yaml"
)

func walkService(name string, v interface{}) *base.ServiceConfig {
	service, ok := v.(map[interface{}]interface{})
	if !ok {
		return nil
	}
	config := &base.ServiceConfig{Name: name}
	val := make(map[string]interface{})
	for k, v := range service {
		if key, ok := k.(string); ok {
			config.SetValueWithKey(key, v)
			val[key] = v
		}
	}
	config.Info = val
	return config
}

func loadConf(path string, prefix string, fun func(*base.ServiceConfig)) error {
	b1, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	m1 := make(map[string]map[string]interface{})
	err = yaml.Unmarshal(b1, &m1)
	if err != nil {
		return err
	}
	services, ok := m1["services"]
	if !ok {
		return fmt.Errorf("services nil")
	}
	for name, v2 := range services {
		if config := walkService(name, v2); config != nil {
			config.Prefix = prefix
			fun(config)
		}
	}
	return nil
}
