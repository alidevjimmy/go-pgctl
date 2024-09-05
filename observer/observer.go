package observer

import (
	"context"
	"encoding/json"
	"time"

	"github.com/alidevjimmy/readyset-replication/config"
	"github.com/alidevjimmy/readyset-replication/election"
	"github.com/alidevjimmy/readyset-replication/node"
	"github.com/alidevjimmy/readyset-replication/zkconn"
	"go.uber.org/zap"
)

// Observer pings nodes and takes action based on the node Role
type Observer struct {
	node       *node.Node
	pool       *node.Pool
	zkc        *zkconn.ZKConnection
	ticker     *time.Ticker
	cancelFunc func()
	ctx        context.Context
	cfg        *config.Config
	logger     *zap.SugaredLogger
}

func NewObserver(n *node.Node, pool *node.Pool, z *zkconn.ZKConnection, interval time.Duration, cfg *config.Config, logger *zap.SugaredLogger) *Observer {
	ctx, cancel := context.WithCancel(context.Background())
	return &Observer{
		node:       n,
		zkc:        z,
		ticker:     time.NewTicker(interval),
		cancelFunc: cancel,
		ctx:        ctx,
		pool:       pool,
		cfg:        cfg,
		logger:     logger,
	}
}

func (o *Observer) Start() {
	o.logger.Infof("observer for node %s started", o.node.ID)
	go o.start()
	run := true
	for run {
		select {
		case <-o.ticker.C:
			o.start()
		case <-o.ctx.Done():
			o.Stop()
			run = false
		}
	}
}

func (o *Observer) start() {
	conn, err := o.node.ConnPool.Acquire(context.Background())
	if err != nil {
		o.logger.Errorf("failed to acquire connection: %v", err)
	}
	defer conn.Release()
	if err := conn.Ping(context.Background()); err != nil {
		switch o.node.Role {
		case node.Leader:
			o.electNewLeader()
			o.pool.RunLeaderQueries()
			o.pool.RunFollowersQueries()
		case node.Follower:
			o.wipeFollower()
		}
	}
}

func (o *Observer) electNewLeader() {
	o.pool.RemoveNode(o.node.ID)
	if len(o.pool.Nodes) == 0 {
		o.logger.Errorf("Node %s is no longer the leader. No more nodes in the pool", o.node.ID)
		return
	}
	leaderIdx := election.Eelect(o.pool.Nodes)
	o.pool.SetLeader(leaderIdx)
	payload, err := json.Marshal(o.pool.Nodes)
	if err != nil {
		o.logger.Errorf("failed to marshal nodes: %v", err)
	}
	o.zkc.Set(o.cfg.NodesPath, payload)
	o.logger.Infof("Node %s is no longer the leader", o.node.ID)
	o.logger.Infof("New leader is %s", o.pool.Nodes[leaderIdx].ID)
	o.cancelFunc()
}

func (o *Observer) wipeFollower() {
	o.pool.RemoveNode(o.node.ID)
	payload, err := json.Marshal(o.pool.Nodes)
	if err != nil {
		o.logger.Errorf("failed to marshal nodes: %v", err)
	}
	o.zkc.Set("o.cfg.NodesPath", payload)
	o.logger.Infof("Node %s is no longer a follower", o.node.ID)
	o.cancelFunc()
}

func (o *Observer) Stop() {
	o.logger.Infof("observer for node %s stopped with err: %s", o.node.ID, o.ctx.Err())
	o.ticker.Stop()
}
