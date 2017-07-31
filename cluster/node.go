package cluster

const (
	MASTER_NODE = 0
	WORKER_NODE = 1
)

type NodeInfo struct {
	Addr string
	Port int
	HostName string
	Role int
}
