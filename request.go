package gorequest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	// "io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type RequestOption struct {
	Method       string
	Url          string
	Head         http.Header
	ConnTimeout  time.Duration
	Timeout      time.Duration
	DelayTime    time.Duration
	JsonFormData *bytes.Reader
	Retrytimes   int
	FormDataType string
}

type Response struct {
	Response *http.Response
	Error    error
}

func Request(url string, method string) *RequestOption {
	// 默认设置
	return &RequestOption{
		Url:         url,
		Method:      method,
		ConnTimeout: 10,
		Timeout:     15,
		Retrytimes:  3,
		DelayTime:   3,
	}
}

func Get(url string) *RequestOption {
	return Request(url, "GET")
}

func Post(url string) *RequestOption {
	return Request(url, "POST")
}

func (option *RequestOption) JsonForm(data *map[string]interface{}) *RequestOption {
	option.FormDataType = "json"
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return option
	}
	option.JsonFormData = bytes.NewReader(bytesData)
	return option
}

func (option *RequestOption) Open() *Response {
	result := &Response{nil, nil}
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				c, err := net.DialTimeout(netw, addr, time.Second*option.ConnTimeout) // 设置建立连接超时
				if err != nil {
					return nil, err
				}
				c.SetDeadline(time.Now().Add(option.Timeout * time.Second)) // 设置发送接收数据超时
				return c, nil
			},
		},
	}

	var req *http.Request
	if option.FormDataType == "json" {
		var err error
		req, err = http.NewRequest(option.Method, option.Url, option.JsonFormData) // 发起请求
		if err != nil {
			result.Error = err
			return result
		}
	}

	// headers
	req.Header.Set("Connection", "close")
	if option.FormDataType == "json" {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// 请求失败重新请求
	var reqerr error
	var resp *http.Response
	for i := 0; i < option.Retrytimes; i++ {
		resp, reqerr = client.Do(req)
		if reqerr != nil {
			fmt.Printf("try again open：%s\r\n", option.Url)
			if option.DelayTime > 0 {
				time.Sleep(option.DelayTime * time.Second)
			}
			continue
		} else {
			break
		}
	}

	if reqerr != nil {
		fmt.Print("Open page fail\n")
		return &Response{resp, nil}
	}

	return &Response{resp, nil}
}

func (req *RequestOption) Retry(n int) *RequestOption {
	req.Retrytimes = n
	return req
}

func (req *RequestOption) Delay(time time.Duration) *RequestOption {
	req.DelayTime = time
	return req
}

// 转码
func (r *Response) Charconv(c string) *Response {
	resp := r.Response
	if r.Error != nil {
		return r
	}
	dec := mahonia.NewDecoder(c)
	resp.Body = ioutil.NopCloser(dec.NewReader(resp.Body))
	return r
}

// match url host
func gethost(url string) string {
	a1 := strings.Split(url, "//")[1]
	return strings.Split(a1, "/")[0]
}

func (response *Response) Request() (*http.Response, error) {
	return response.Response, response.Error
}

func (response *Response) Json(t interface{}) error {
	resp := response.Response
	if response.Error != nil {
		return response.Error
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, t)
	return err
}

// jsonp to json
// func (response *Response) Jsonp() (*simplejson.Json, error) {
// 	defer response.Response.Body.Close()
// 	body, _ := ioutil.ReadAll(response.Response.Body)

// 	// match json
// 	reg := regexp.MustCompile(`^[^\[{]*([\[{][\s\S]*?[\]}])[^\]}]*$`)
// 	match := reg.FindSubmatch(body)
// 	if len(match) < 2 {
// 		return nil, errors.New("jsonp to json error")
// 	}

// 	return simplejson.NewJson(match[1])
// }

func (response *Response) Html() (*goquery.Document, error) {
	if response.Error != nil {
		return nil, response.Error
	}
	return goquery.NewDocumentFromResponse(response.Response)
}
