package server

import (
	"github.com/wooddeep/gbat/worker/client"
	"github.com/wooddeep/gbat/worker/action"
	"github.com/Jeffail/gabs"
	"net/http"
	"os/exec"
	"strconv"
	"unsafe"
	"regexp"
	"bytes"
	"net"
	"fmt"
	"os"
	"io"
)

var uploadDir = "./"

func HelloWorld(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hello world"))
}

var UploadHandle = func(w http.ResponseWriter, r *http.Request) {
	file, head, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	re := regexp.MustCompile(`([:\w\d\\\/]+[\\\/])([^\\\/]+)`)
	match := re.FindStringSubmatch(head.Filename)
	exe := uploadDir + match[len(match)-1]
	fW, err := os.Create(exe)
	if err != nil {
		fmt.Println("文件创建失败")
		return
	}
	defer func() {
		fW.Close()
		go func() {
			fmt.Println("#######" + exe)
			//arg := []string{`F:\work\go\test\agt00`}
			// TOOD　添加执行进度条
			cmd := exec.Command(exe, "C:\\Users\\lihan\\Desktop\\cxl\\python\\xx") // TODO 替换待分析的目录
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				return
			}

			fmt.Printf("callEXE2结果:\n%v\n\n%v\n\n%v", string(output), cmd.Stdout, cmd.Stderr)
			fmt.Println("end #######")
		}()
	}()
	_, err = io.Copy(fW, file)
	if err != nil {
		fmt.Println("文件保存失败")
		return
	}
	fmt.Fprintf(w, "#save file ok")
}

var CollectorHandler = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	colladdr := r.URL.Query().Get("colladdr")
	//fmt.Println("##colladdr", colladdr)
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	fmt.Fprintf(w, `{"code":0, "msg":"success!"}`)

	coll, _ := strconv.ParseUint(colladdr, 0, 64) // 字符串转整形
	pcoll := (*uint64)(unsafe.Pointer(uintptr(coll)))  // 整形转指针
	pch := (*chan string)(unsafe.Pointer(pcoll))    // 整形指针转channel指针
	*pch <- string(body)

}

// curl -X GET http://127.0.0.1:55167/start -d '{"name":"lihan"}'
// fs.dir("F:\\log").find('(.acd=="20000731") && (.endTime=="20170701235958545")').exe()
var WorkHandler = func(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	len := r.ContentLength
	body := make([]byte, len)
	r.Body.Read(body)
	obj, err := gabs.ParseJSON([]byte(body))
	if err != nil {
		fmt.Fprintf(w, `{"code": 1, "msg":"json parse error!", "data":{}}`)
		return
	}
	fmt.Fprintf(w, string(body))

	colladdr, _ := obj.Path("colladdr").Data().(string)
	query, _ := obj.Path("query").Data().(string)
	path, _ := obj.Path("path").Data().(string)
	url, _ := obj.Path("url").Data().(string)
	buffer := bytes.Buffer{}
	collector := make(chan string)
	notifier := make(chan bool)

	go action.FileParseMain(notifier, collector, path, query)
	go func(buff bytes.Buffer, coll chan string, notify chan bool) {
		for {
			exitFlag := false
			select {
			case str := <-collector:
				buffer.WriteString(str)
				buffer.WriteString(`,`)
			case exitFlag = <-notifier:
				fmt.Println(`exit`)
			}

			if exitFlag {
				break
			}
		}
		buffer.WriteString(`{}`)
		err = client.HttpPost("http://" + url + "/report?colladdr=" + colladdr, buffer.String()) // 发送消息给执行节点

	}(buffer, collector, notifier)
}

func StartHttpServ() net.Listener {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
		return nil
	}

	http.HandleFunc("/", HelloWorld)
	http.HandleFunc("/report", CollectorHandler)
	http.HandleFunc("/start", WorkHandler)

	go func() {
		http.Serve(listener, nil)
	}()

	go func() {
		port := listener.Addr().(*net.TCPAddr).Port
		os.Mkdir("./"+strconv.Itoa(port), 0777)
		uploadDir = "./" + strconv.Itoa(port) // TODO 修改文件创建方式
	}()

	return listener
}
