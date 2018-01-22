package gorequest

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
	// "bytes"
	"errors"
	"github.com/PuerkitoBio/goquery"
	"github.com/axgle/mahonia"
	"github.com/bitly/go-simplejson"
	// "io"
	"net/url"
	"regexp"
	"strings"
)

type RequestOption struct {
	Method      string
	Url         string
	Head        http.Header
	ConnTimeout time.Duration
	Timeout     time.Duration
	DelayTime   time.Duration
	Data        url.Values
	Retrytimes  int
}

type Response struct {
	Response *http.Response
	Error    error
}

// simple downloader
func Download(url string) *Response {
	client := &http.Client{}

	req, _ := http.NewRequest("GET", url, nil) // 发起请求

	host := gethost(url) // get host

	req.Header.Add("Host", host)
	req.Header.Add("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/62.0.3202.94 Safari/537.36")

	// 请求失败重新请求
	var reqerr error
	var resp *http.Response
	retry := 3 // 重新请求次数
	for i := 0; i < retry; i++ {
		resp, reqerr = client.Do(req)
		if reqerr != nil {
			continue
		}
	}
	if reqerr != nil {
		fmt.Println("网页打开失败")
		panic(reqerr)
	}

	response := Response{
		Response: resp,
	}

	return &response
}

func (option *RequestOption) Download() *Response {
	result := &Response{nil, nil} // 生命返回变量
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

	// form
	formio := option.Data.Encode()
	form := strings.NewReader(formio)

	req, err := http.NewRequest(option.Method, option.Url, form) // 发起请求
	if err != nil {
		result.Error = err
		return result
	}

	// header
	// req.Header = req.Head
	req.Header.Set("Connection", "close")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// 请求失败重新请求
	var reqerr error
	var resp *http.Response
	for i := 0; i < option.Retrytimes; i++ {
		resp, reqerr = client.Do(req)
		if reqerr != nil {
			fmt.Printf("请求失败重试：%d\r\n", i+1)
			if option.DelayTime > 0 {
				time.Sleep(option.DelayTime * time.Second)
			}
			continue
		} else {
			break
		}
	}

	if reqerr != nil {
		fmt.Print("网页打开失败\n")
		panic(reqerr)
	}

	return &Response{resp, nil}
}

// 转码
func (r *Response) Charconv(c string) *Response {
	resp := r.Response
	dec := mahonia.NewDecoder(c)
	resp.Body = ioutil.NopCloser(dec.NewReader(resp.Body))
	return r
}

// 获取url的host
func gethost(url string) string {
	a1 := strings.Split(url, "//")[1]
	return strings.Split(a1, "/")[0]
}

func (response *Response) Bytes() ([]byte, error) {
	defer response.Response.Body.Close()
	return ioutil.ReadAll(response.Response.Body)
}

func (response *Response) Json() (*simplejson.Json, error) {
	defer response.Response.Body.Close()
	return simplejson.NewFromReader(response.Response.Body)
}

/*
	jsonp 转 json
	正则提取出jsonp的json
*/
func (response *Response) Jsonp() (*simplejson.Json, error) {
	defer response.Response.Body.Close()
	body, _ := ioutil.ReadAll(response.Response.Body)

	reg := regexp.MustCompile(`^[^\[{]*([\[{][\s\S]*?[\]}])[^\]}]*$`) // 提取json正则表达式
	match := reg.FindSubmatch(body)                                   // 提取json
	if len(match) < 2 {
		return nil, errors.New("jsonp提取json失败，正则无法匹配")
	}

	return simplejson.NewJson(match[1])
}

func (response *Response) Html() (*goquery.Document, error) {
	if response.Error != nil {
		return nil, response.Error
	}
	if response.Response == nil {
		return nil, errors.New("response is nil")
	}
	return goquery.NewDocumentFromResponse(response.Response)
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

func (req *RequestOption) Retry(n int) *RequestOption {
	req.Retrytimes = n
	return req
}

func (req *RequestOption) Delay(time time.Duration) *RequestOption {
	req.DelayTime = time
	return req
}
