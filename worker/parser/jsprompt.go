package parser

import (
	//"github.com/robertkrimen/otto"
	//"gbat/worker"
	//"fmt"
	//"os"
)

//Set a Go function
//
//vm.Set("sayHello", func(call otto.FunctionCall) otto.Value {
//	fmt.Printf("Hello, %s.\n", call.Argument(0).String())
//	return otto.Value{}
//})
//Set a Go function that returns something useful
//
//vm.Set("twoPlus", func(call otto.FunctionCall) otto.Value {
//	right, _ := call.Argument(0).ToInteger()
//	result, _ := vm.ToValue(2 + right)
//	return result
//})
//Use the functions in JavaScript
//
//result, _ = vm.Run(`
//    sayHello("Xyzzy");      // Hello, Xyzzy.
//    sayHello();             // Hello, undefined
//
//    result = twoPlus(2.0); // 4
//`)

// fs.find({"req.oldt.$.spc":"698042"})
// fs.find({"req.oldt.$.spc":"698042"})

// db.tianyc02.find({$or:[{age:11},{age:22}]})

var Prompt = `

var fs = {
	path  : ".",
	query : "."
};

fs.dir = function(path) {
	if (path == undefined) path = ".";
	fs.path = path;
	return fs;
}

fs.find = function(query) {
	if (query == undefined) query = ".";
	//fs.query = JSON.stringify(query);
	fs.query = query
	return fs;
}

fs.exe = function() {
	out = Goprompt(fs.path, fs.query); // go function
	return eval(out);
}


`
/*
 * called by js
 *	fmt.Printf("para[0]: %s\n", call.Argument(0).String())
 *	fmt.Printf("para[1]: %s\n", call.Argument(1).String())
 */
//func Goprompt(call otto.FunctionCall) otto.Value {
//
//	path := call.Argument(0).String()
//	query := call.Argument(1).String()
//
//	_, err := os.Stat(path)
//	if err != nil {
//		fmt.Printf("# path <%s> not found!\n", path);
//		return otto.Value{}
//	}
//
//	go worker.FileParseMain(path, query)
//
//	return otto.Value{}
//
//}