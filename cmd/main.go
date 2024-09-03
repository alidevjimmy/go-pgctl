package main

import (
	"log"

	"github.com/alidevjimmy/readyset-replication/config"
)

func main() {
	// get node from config file
	cfg, err := config.New("./config/config.yaml")
	if err != nil {
		log.Panicf("failed to read config file: %v", err)
	}
	log.Println(cfg.Pool.Nodes)
	// connect to zookeeper

	// connect to nodes

	// add node to zookeeper

	// elect one of nodes in zookeeper

	// run leader commands on leader node

	// run follower commands on follower node

	// watch node priodically to check if nodes is still alive

	// if leader node fails, elect new leader

	// run leader commands on new leader node
}
