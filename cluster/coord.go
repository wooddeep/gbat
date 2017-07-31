package cluster

import (
	"strconv"
	"time"
	"log"
	"fmt"
)

//https://acupple.github.io/2016/05/31/ETCD%E5%AE%9E%E7%8E%B0leader%E9%80%89%E4%B8%BE/

//func (coord *EtcdCoord) Register(remoteAddr string, remotePort int) chan NodeInfo {
//	cfg := client.Config{
//		Endpoints: []string{"http://127.0.0.1:2379"},
//		Transport: client.DefaultTransport,
//		// set timeout per request to fail fast when the target endpoint is unavailable
//		HeaderTimeoutPerRequest: time.Second,
//	}
//	c, err := client.New(cfg)
//	if err != nil {
//		log.Fatal(err)
//	}
//	kapi := client.NewKeysAPI(c)
//	// set "/foo" key with "bar" value
//	//log.Print("Setting '/foo' key with 'bar' value")
//	//opt := &client.SetOptions{}
//	//opt.TTL = 100000
//
//	// to get the master id registerd in the etcd
//	resp, err := kapi.Get(context.Background(), "/deer", nil)
//	if err == nil {
//		// select a random port as the master's communication port
//
//		innresp, innerr := kapi.Set(context.Background(), "/deer", "bar", nil)
//
//		_ = innresp
//		_ = innerr
//
//		log.Print(resp.Node.Key, resp.Node.Value)
//	} else {
//
//	}
//
//	resp, err = kapi.Set(context.Background(), "/foo", "bar", nil)
//	if err != nil {
//		log.Fatal(err)
//	} else {
//		// print common key info
//		log.Printf("Set is done. Metadata is %q\n", resp)
//	}
//	// get "/foo" key's value
//	log.Print("Getting '/foo' key value")
//	resp, err = kapi.Get(context.Background(), "/foo", nil)
//	if err != nil {
//		log.Fatal(err)
//	} else {
//		// print common key info
//		log.Printf("Get is done. Metadata is %q\n", resp)
//		// print value
//		log.Printf("%q key has %q value\n", resp.Node.Key, resp.Node.Value)
//	}
//	return nil
//}

//////////////////////////////////////////////////////////////////////////////////////////////////////////

//type EtcdCoord struct {
//}
//var kapi client.KeysAPI = nil
//
//func (coord *EtcdCoord) InitKapi(etcdAddr string, etcdPort int) {
//	cfg := client.Config{
//		Endpoints:               []string{"http://" + etcdAddr + ":" + strconv.Itoa(etcdPort)},
//		Transport:               client.DefaultTransport,
//		HeaderTimeoutPerRequest: time.Second,
//	}
//
//	c, err := client.New(cfg)
//	if err != nil {
//		panic(err)
//	}
//
//	kapi = client.NewKeysAPI(c)
//}
//
//func (coord *EtcdCoord) Keepalive(key, val string, ts time.Duration) (chan int) {
//	ticker := time.NewTicker((ts - 1) * time.Second)
//	quit := make(chan int)
//	opts := &client.SetOptions{}
//	opts.TTL = time.Second * ts
//
//	//<-notifier
//
//	go func() {
//		for {
//			select {
//			case <-ticker.C:
//				kapi.Set(context.Background(), key, val, opts)
//			case <-quit:
//				ticker.Stop()
//				return
//			}
//		}
//	}()
//
//	return quit
//}
//
//func KeepDominate(key, val string, ts time.Duration) (chan int) {
//	ticker := time.NewTicker((ts - 1) * time.Second)
//	quit := make(chan int)
//	opts := &client.SetOptions{}
//	opts.TTL = time.Second * ts
//
//	go func() {
//		for {
//			select {
//			case <-ticker.C:
//				kapi.Set(context.Background(), key, val, opts)
//			case <-quit:
//				ticker.Stop()
//				return
//			}
//		}
//	}()
//
//	return quit
//}
//
//func (coord *EtcdCoord) WatchDeer(localAddr, etcdAddr string, localPort, etcdPort int) {
//	go func() {
//		id := localAddr + ":" + strconv.Itoa(localPort)
//		watcher := kapi.Watcher("/deer", &client.WatcherOptions{
//			Recursive: false,
//		})
//
//		for {
//			res, err := watcher.Next(context.Background())
//			if err != nil {
//				log.Println("Error watch workers:", err)
//				break
//			}
//
//			if res.Action == "expire" {
//				log.Printf("# I am %s, I will hunt the deer!", id)
//				coord.Register(localAddr, etcdAddr, localPort, etcdPort)
//			} else if res.Action == "set" || res.Action == "update" {
//				// pass
//			} else if res.Action == "delete" {
//				// pass
//			}
//		}
//	} ()
//}
//
//func (coord *EtcdCoord) Register(localAddr, etcdAddr string, localPort, etcdPort int) {
//
//	// to get the master id registerd in the etcd
//	winner := ""
//	go func() {
//		resp, err := kapi.Get(context.Background(), "/deer", nil)
//		opts := &client.SetOptions{}
//		opts.TTL = time.Second * 10
//		if err != nil { // not found
//			innresp, innerr := kapi.Set(context.Background(), "/deer", localAddr+":"+strconv.Itoa(localPort), opts)
//			if innerr != nil {
//				panic("# catch deer fail!")
//			}
//
//			winner = innresp.Node.Value
//
//			// keep the winner
//			log.Print("# yes I am the winner, I will keep the deer!")
//			KeepDominate("/deer", localAddr+":"+strconv.Itoa(localPort), 5)
//
//		} else { // found ! as a loser I will watch the deer
//			log.Print("# yes I am a loser, I will watch the deer!")
//			winner = resp.Node.Value
//			coord.WatchDeer(localAddr, etcdAddr, localPort, etcdPort)
//		}
//
//		//notifier <- true
//	} ()
//
//}
// use age for EtcdCoord
//etcdCoord := &cluster.EtcdCoord{}
//etcdCoord.InitKapi(etcdip, etcdpo)
//etcdCoord.Register(lip, etcdip, port, etcdpo);
//etcdCoord.Keepalive("/heros/" + lip + ":" + strconv.Itoa(port), "", 5)


var coordApi ICoordApi = &EtcdCoordApi{}

func InitApi(coordAddr string, coordPort int) {
	coordApi.Init(coordAddr, coordPort)
}

func Keepalive(key, val string, ts time.Duration) (chan int) {
	ticker := time.NewTicker((ts - 1) * time.Second)
	quit := make(chan int)

	para := CoordPara{}
	para.TTL = time.Second * ts

	go func() {
		for {
			select {
			case <-ticker.C:
				coordApi.Set(key, val, para)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit
}

func KeepWinner(key, val string, ts time.Duration) (chan int) {
	ticker := time.NewTicker((ts - 1) * time.Second)
	quit := make(chan int)

	para := CoordPara{}
	para.TTL = time.Second * ts

	go func() {
		for {
			select {
			case <-ticker.C:
				coordApi.Set(key, val, para)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	return quit
}

func WatchDeer(localAddr, etcdAddr string, localPort, etcdPort int) {
	log.Println("## WatchDeer")
	go func() {
		watcher := coordApi.Watch("/deer", CoordPara{})
		para := CoordPara{}
		para.localAddr = localAddr
		para.etcdAddr = etcdAddr
		para.localPort = localPort
		para.etcdPort = etcdPort
		coordApi.WatchHandler(watcher, para, Register)
	} ()
}

func Register(localAddr, etcdAddr string, localPort, etcdPort int) {
	//winner := ""
	go func() {
		_, err := coordApi.Get("/deer", CoordPara{})
		if err != nil { // not found
			para := CoordPara{}
			para.TTL = time.Second * 10
			innresp, innerr := coordApi.Set("/deer", localAddr + ":" + strconv.Itoa(localPort), para)
			if innerr != nil {
				panic("# catch deer fail!")
			}
			_ = innresp
			//winner = innresp.Node.Value

			// keep the winner
			log.Print("# yes I am the winner, I will keep the deer!")
			KeepWinner("/deer", localAddr + ":" + strconv.Itoa(localPort), 5)

		} else { // found ! as a loser I will watch the deer
			log.Print("# yes I am a loser, I will watch the deer!")
			//winner = resp.Node.Value
			WatchDeer(localAddr, etcdAddr, localPort, etcdPort)
		}

	} ()
}

/*
## v: /heros/127.0.0.1:64192
## v: /heros/127.0.0.1:64176
*/
func GetWokerList() []string {
	resp, _ := coordApi.GetDir("/heros", CoordPara{})
	_ = resp
	for _, v := range resp {
		fmt.Printf("## v: %v\n", v)
	}
	return resp
}