package parser

import (
	"github.com/xwb1989/sqlparser"
	"fmt"
)

func _main () {
	ret, _ := sqlparser.Parse("select :a from a where a in (:b)")
	fmt.Println(ret)
}