package main

import (
	"github.com/wooddeep/gbat/worker/parser"
	"github.com/wooddeep/gbat/worker/server"
	"github.com/wooddeep/gbat/worker/client"
	"github.com/wooddeep/gbat/worker/action"
	"github.com/wooddeep/gbat/cluster"
	"github.com/wooddeep/gbat/utils"
	"github.com/robertkrimen/otto/repl"
	"github.com/robertkrimen/otto"
	"github.com/Jeffail/gabs"
	"os/signal"
	"syscall"
	"strings"
	"strconv"
	"regexp"
	"bytes"
	"flag"
	"fmt"
	"net"
	"os"
)

const (
	STANDALONE_NODE = 0
	CLUSTER_NODE_WITH_SHELL = 1
	CLUSTER_NODE_WITHOUT_SHELL = 2
)

var workMode = STANDALONE_NODE // 0 ~ standalone, 1 ~ node with shell, 2 ~ node without shell
var autoPort = 0
var localIp  = ""

/*
 * gbat --etcd 127.0.0.1:2379
 * gbat --etcd 127.0.0.1:2379 --commit f:/work/xx.exe
 */
func startRepl () {
	vm := otto.New()
	vm.Run(parser.Prompt)
	vm.Set("exit", func(call otto.FunctionCall) otto.Value {
		os.RemoveAll("./" + strconv.Itoa(autoPort))
		os.Exit(0)
		return otto.Value{}
	})

	vm.Set("lsnode", func(call otto.FunctionCall) otto.Value {
		if workMode == STANDALONE_NODE {
			return otto.Value{}
		}

		cluster.GetWokerList()
		return otto.Value{}
	})

	vm.Set("Goprompt", func(call otto.FunctionCall) otto.Value {
		path := call.Argument(0).String()
		query := call.Argument(1).String()
		_, err := os.Stat(path)
		if err != nil {
			fmt.Printf("# path <%s> not found!\n", path);
			return otto.Value{}
		}

		if workMode == CLUSTER_NODE_WITH_SHELL {
			workers := cluster.GetWokerList() // 查询集群节点 通知各个节点查询 //v: /heros/127.0.0.1:56126
			collector := make(chan string, len(workers))
			//fmt.Printf("## collector addr : %v\n", &collector)
			colladdr := fmt.Sprintf("%v", &collector) // 获取collector的地址转化为字符串
			for _, work := range workers {
				re := regexp.MustCompile(`heros.([0-9:\.]+)`)
				match := re.FindStringSubmatch(work)
				if match != nil {
					url := match[1]
					body := gabs.New()
					body.Set(path, "path")
					body.Set(query, "query")
					body.Set(colladdr, "colladdr") // 把collector的地址序列化, 传到每一个worker
					body.Set(localIp + ":" + strconv.Itoa(autoPort), "url")
					client.HttpPost("http://" + url + "/start", body.String())
				}
			}
			buffer := bytes.Buffer{}
			buffer.WriteString(`[`)

			for i := 0; i < len(workers); i++ {
				ret := <- collector
				buffer.WriteString(ret)
				buffer.WriteString(`,`)
			}

			buffer.WriteString(`{}]`)
			out, _ := vm.ToValue(buffer.String())
			//fmt.Println("## out \n", buffer.String())

			return out
		}

		if workMode == STANDALONE_NODE {
			buffer := bytes.Buffer{}
			buffer.WriteString(`[`)
			collector := make(chan string)
			notifier := make(chan bool)
			go action.FileParseMain(notifier, collector, path, query)
			for {
				exitFlag := false
				select {
				case str := <-collector:
					fmt.Println(str)
					buffer.WriteString(str)
					buffer.WriteString(`,`)
				case exitFlag = <-notifier:
					fmt.Println(`exit`)
				}

				if exitFlag {
					break
				}
			}

			buffer.WriteString(`{}]`)
			out, _ := vm.ToValue(buffer.String())
			return out
		}

		return otto.Value{}
	})

	if err := repl.Run(vm); err != nil {
		panic(err)
	}
}

func cleanup() {
	os.RemoveAll("./" + strconv.Itoa(autoPort))
	os.Exit(0)
}

var etcdaddr string

func init() {
	flag.StringVar(&etcdaddr, "etcdaddr", "127.0.0.1:2379", "help message for flagname")
	//flag.Var(&shell, "shell", "xxxx")
}

func main() {
	//flag.Parse()
	argn := len(os.Args)
	if argn == 1 { // standalone mode!
		startRepl()
		return
	}

	rip, rport, shell := "", 0, ""
	args := strings.Join(os.Args, " ")
	re := regexp.MustCompile(`\-\-etcd\s+([\d\.]+):(\d+)\s?(\-\-shell)?\s?`)
	match := re.FindStringSubmatch(args)
	if match != nil {
		rip = match[1]
		rport, _ = strconv.Atoi(match[2])
		shell = match[3]
	} else {
		fmt.Println("sub parameter error, usage:")
		fmt.Println("\tgbat --etcd 127.0.0.1:2379")
		fmt.Println("\tgbat --etcd 127.0.0.1:2379 --shell")
		return
	}

	workMode = CLUSTER_NODE_WITHOUT_SHELL
	if strings.Count(shell, "") > 1 { // shell mode
		workMode = CLUSTER_NODE_WITH_SHELL
		go startRepl()
	}

	cluster.InitApi(rip, rport)
	lip, err := utils.GetLocalIp(rip)
	if err != nil {
		lip = "127.0.0.1"
		localIp = lip
	}

	lsn := server.StartHttpServ()          // start the http server
	port := lsn.Addr().(*net.TCPAddr).Port // get the dynamic generated http port
	autoPort = port

	cluster.Register(lip, rip, port, rport)
	cluster.Keepalive("/heros/"+lip+":"+strconv.Itoa(port), "", 5)

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		cleanup()
		os.Exit(1)
	}()

	c := make(chan int)
	<-c // wait forerver


}
