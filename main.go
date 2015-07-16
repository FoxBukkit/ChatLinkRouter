package main

import (
	"encoding/json"
	zmq "github.com/pebbe/zmq4"
	"log"
	"os"
)

type ZMQConfig struct {
	Uri  string
	Type string
}

type ZMQConfigFile struct {
	ServerToMe []ZMQConfig
	MeToLink   []ZMQConfig
	LinkToMe   []ZMQConfig
	MeToServer []ZMQConfig
}

func applyZMQConfig(socket *zmq.Socket, configs []ZMQConfig) {
	for _, config := range configs {
		switch config.Type {
		case "bind":
			socket.Bind(config.Uri)
		case "connect":
			socket.Connect(config.Uri)
		}
	}
}

func main() {
	file := new(ZMQConfigFile)

	fileReader, err := os.Open("config.json")
	if err != nil {
		log.Panicf("Error: %v", err)
	}
	jsonReader := json.NewDecoder(fileReader)
	err = jsonReader.Decode(&file)
	fileReader.Close()
	if err != nil {
		log.Panicf("Error: %v", err)
	}

	serverToMe, _ := zmq.NewSocket(zmq.PULL)
	defer serverToMe.Close()
	applyZMQConfig(serverToMe, file.ServerToMe)
	meToLink, _ := zmq.NewSocket(zmq.PUSH)
	defer meToLink.Close()
	applyZMQConfig(meToLink, file.MeToLink)

	linkToMe, _ := zmq.NewSocket(zmq.XSUB)
	defer linkToMe.Close()
	applyZMQConfig(linkToMe, file.LinkToMe)
	meToServer, _ := zmq.NewSocket(zmq.XPUB)
	defer meToServer.Close()
	applyZMQConfig(meToServer, file.MeToServer)

	go zmq.Proxy(serverToMe, meToLink, nil)
	zmq.Proxy(linkToMe, meToServer, nil)
}
