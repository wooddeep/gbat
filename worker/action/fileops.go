package action

import (
	"github.com/chrislusf/glow/flow"
	"github.com/nytlabs/gojee"
	"encoding/json"
	"path/filepath"
	"sync"
	"fmt"
	"os"
)

//func wait() {
//	timer := time.NewTicker(3 * time.Second)
//	<-timer.C
//}

//fs.dir("F:\\log").find('(.acd=="20000731") && (.endTime=="20170701235958545")').exe()
func parseFilelist(coll chan string, w *sync.WaitGroup, c chan int, path string, query *jee.TokenTree) {
	err := filepath.Walk(path, func (path string, f os.FileInfo, err error) error {
		if (f == nil) {return err}
		if f.IsDir() {return nil}

		w.Add(1)
		go func(coll chan string, waiter *sync.WaitGroup){
			doFileParse(coll, w, c, path, query)
		}(coll, w)

		return nil
	})

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err)
	}
}

func Match(line string, tree *jee.TokenTree) bool {
	var umsg jee.BMsg
	err := json.Unmarshal([]byte(line), &umsg)
	if err != nil {
		return false
	}

	result, err := jee.Eval(tree, umsg)
	if err != nil {
		return false
	}

	switch result.(type) {
	case bool:
		b, _ := result.(bool)
		return b
	}

	return false
}

/*
	flow.New().TextFile(file, 3,
	).Filter(func(line string) bool {
		return Match(line, query)
	}).Map(func(line string, ch chan string) {
		for _, token := range strings.Split(line, ":") {
			ch <- token
		}
	}).Map(func(key string) int {
		return 1
	}).Reduce(func(x int, y int) int {
		return x + y
	}).Map(func(x int) {
		println("file:", file, "\t","count:", x)
	}).Run()
*/

func doFileParse(coll chan string, waiter *sync.WaitGroup, c chan int, file string, query *jee.TokenTree) {
	c <- 0

	flow.New().TextFile(file, 10,
	).Filter(func(line string) bool {
		return Match(line, query)
	}).Map(func(line string) {
		//log.Printf("# line: %s", line)
		coll <- line
	}).Run()

	waiter.Done()
	<- c
}

func FileParseMain(notifier chan bool, coll chan string, path, query string) {
	l, err := jee.Lexer(query)
	if err != nil {
		notifier <- true
		return
	}

	tree, err := jee.Parser(l)
	if err != nil {
		notifier <- true
		return
	}

	waiter := &sync.WaitGroup{}
	scheduler := make(chan int, 10)
	parseFilelist(coll, waiter, scheduler, path, tree)
	waiter.Wait()

	fmt.Println("## yes I have complete the log parse!")

	notifier <- true
}


