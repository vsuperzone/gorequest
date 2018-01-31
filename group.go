package gorequest

import (
// "fmt"
)

type GroupData []string

func Group(urls []string) *GroupData {
	var data GroupData
	data = urls
	return &data
}

func (urls *GroupData) Each(n int, f func(string)) {
	list := make(chan bool, n)
	for _, url := range *urls {
		go func() {
			list <- true
			f(url)
		}()
		<-list
	}
	close(list)
}
