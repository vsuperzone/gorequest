package gorequest

import (
	"fmt"
	"testing"
)

func Test_get(t *testing.T) {
	bytes, _ := Get("http://wl.localhost/api/yunwangke/partner/all").Download().Bytes()
	fmt.Println(bytes)
}
