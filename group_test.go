package gorequest

import (
	"fmt"
	"testing"
)

func Test_group(t *testing.T) {
	urls := []string{
		"http://baidu.com",
		"http://163.com",
	}
	Group(urls)
}

func Test_groupget(t *testing.T) {
	urls := []string{
		"http://baidu.com",
		"http://163.com",
	}

	Group(urls).Each(10, func(url string) {
		resp := Get(url).Open()
		fmt.Println(resp)
	})
}
