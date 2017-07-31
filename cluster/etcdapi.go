package cluster

import (
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"strconv"
	"time"
	"log"
	//"fmt"
)

type CoordPara struct {
	TTL time.Duration
	localAddr string
	etcdAddr string
	localPort int
	etcdPort int
}

type EtcdCoordApi struct {
	kapi client.KeysAPI
}

func (coord *EtcdCoordApi) Init(coordAddr string, coordPort int) {

	cfg := client.Config{
		Endpoints:               []string{"http://" + coordAddr + ":" + strconv.Itoa(coordPort)},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}

	c, err := client.New(cfg)
	if err != nil {
		panic(err)
	}

	coord.kapi = client.NewKeysAPI(c)
}

func (coord* EtcdCoordApi) Set(key, val string, para CoordPara) (interface{}, error) {
	opts := &client.SetOptions{}
	opts.TTL = para.TTL
	return coord.kapi.Set(context.Background(), key, val, opts)
}

func (coord* EtcdCoordApi) Get(key string, para CoordPara) (interface{}, error) {
	return coord.kapi.Get(context.Background(), key, nil)
}

func (coord* EtcdCoordApi) GetDir(key string, para CoordPara) ([]string, error) {
	out := [] string {}
	resp, err := coord.kapi.Get(context.Background(), key, nil)
	if err != nil {
		return nil, err
	}

	nodes := resp.Node.Nodes
	for _, node:= range nodes {
		out = append(out, node.Key)
	}

	return out, nil
}

func (coord* EtcdCoordApi) Watch(key string, para CoordPara) (interface{}) {
	watcher := coord.kapi.Watcher(key, &client.WatcherOptions{
		Recursive: false,
	})
	return watcher
}

func (coord* EtcdCoordApi) WatchHandler(w interface{}, para CoordPara, reg func(string, string, int, int)) {
	for {
		watcher, _ := w.(client.Watcher)

		res, err := watcher.Next(context.Background())
		if err != nil {
			log.Println("Error watch workers:", err)
			break
		}

		id := para.localAddr + ":" + strconv.Itoa(para.localPort)
		if res.Action == "expire" {
			log.Printf("# I am %s, I will hunt the deer!", id)
			reg(para.localAddr, para.etcdAddr, para.localPort, para.etcdPort)
			break
		} else if res.Action == "set" || res.Action == "update" {
			// pass
		} else if res.Action == "delete" {
			// pass
		}
	}
}