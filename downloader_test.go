package gorequest

import (
	"fmt"
	"testing"
)

func Test_get(t *testing.T) {
	fmt.Println(Get("http://baidu.com"))
}
