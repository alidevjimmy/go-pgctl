package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alidevjimmy/readyset-replication/config"
	"github.com/alidevjimmy/readyset-replication/election"
	"github.com/alidevjimmy/readyset-replication/node"
	"github.com/alidevjimmy/readyset-replication/observer"
	"github.com/alidevjimmy/readyset-replication/readyset"
	"github.com/alidevjimmy/readyset-replication/zkconn"
	"go.uber.org/zap"
)

var (
	pool  *node.Pool
	zkc   *zkconn.ZKConnection
	sugar *zap.SugaredLogger
)

func main() {

	// init logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	sugar = logger.Sugar()

	// get node from config file
	cfg, err := config.New("./config/config.yaml")
	if err != nil {
		sugar.Fatalf("failed to read config file: %v", err)
	}

	// connect to nodes
	nodes, err := initNodes(cfg)
	if err != nil {
		sugar.Errorf("failed to connect to nodes: %v", err)
	}

	pool = node.NewPool(nodes, sugar)

	// connect to zookeeper
	zkc, err = zkconn.Connect(cfg.Zookeeper, 5*time.Second)
	if err != nil {
		sugar.Fatalf("failed to connect to zookeeper: %v", err)
	}
	// elect one of nodes as leader
	leaderIdx := election.Eelect(nodes)
	pool.SetLeader(leaderIdx)
	sugar.Infof("Node %s is the leader", nodes[leaderIdx].ID)

	// add nodes to zookeeper
	payload, err := json.Marshal(nodes)
	if err != nil {
		sugar.Fatalf("failed to marshal nodes: %v", err)
	}
	path := cfg.NodesPath
	exists, err := zkc.Exists(path)
	if err != nil {
		sugar.Fatalf("failed to check if node exists: %v", err)
	}
	if !exists {
		if err := zkc.Create(path, payload); err != nil {
			sugar.Fatalf("failed to create node: %v", err)
		}
	} else {
		if err := zkc.Set(path, payload); err != nil {
			sugar.Fatalf("failed to set node: %v", err)
		}
	}
	sugar.Infof("Zookeeper is in-sync")

	// run node observers for each node
	for _, n := range nodes {
		obs := observer.NewObserver(n, pool, zkc, 5*time.Second, cfg, sugar)
		go obs.Start()
	}

	// run leader commands on leader node
	pool.RunLeaderQueries()
	// run follower commands on follower node
	pool.RunFollowersQueries()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)
	sugar.Infof("Cought signal: %s, terminating...", <-sigChan)
}

func initNodes(cfg *config.Config) ([]*node.Node, error) {
	nodes := make([]*node.Node, 0)
	for _, n := range cfg.Pool.Nodes {
		rs, err := readyset.NewRS(n.DSN)
		if err != nil {
			sugar.Errorf("failed to connect to readyset: %v", err)
			return nil, err
		}
		nodes = append(nodes, node.NewNode(rs.ConnPool, n.ID, n.DSN, n.InternalHost, n.InternalPort))
	}
	return nodes, nil
}
