package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"strconv"
)

type Endpoint struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
	User string `yaml:"user"`
}

var endpoints = []Endpoint{}

// 	{
// 		Host:     "37.187.238.132",
// 		Port:     "22",
// 		User:     "wayt",
// 		Password: "toto4242",
// 	},
// }

func loadEndpoints(filepath string) error {

	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &endpoints); err != nil {
		return err
	}

	log.Println("loaded ", len(endpoints), "endpoint(s)")
	return nil
}

func getEndpoint(i string) (Endpoint, bool) {

	num, err := strconv.ParseInt(i, 10, 32)
	if err != nil {
		return Endpoint{}, false
	}

	if num < 0 || int(num) >= len(endpoints) {
		return Endpoint{}, false
	}

	return endpoints[int(num)], true
}
