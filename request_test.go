package gorequest

import (
	"fmt"
	"io/ioutil"
	"testing"
)

func Test_get(t *testing.T) {
}

func Test_post(t *testing.T) {
	data := make(map[string]interface{})
	data["name"] = "李白"
	data["timelength"] = 128
	data["author"] = "李荣浩"
	url := "http://wl.localhost/api/rankquery/add_task"
	resp, _ := Post(url).JsonForm(&data).Open().Request()
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
}
