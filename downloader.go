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
	"encoding/json"
	"net/url"
	"regexp"
	"strings"
)

type RequestOption struct {
	Method      string // post or get...
	Url         string
	Head        http.Header
	ConnTimeout time.Duration // 连接超时时间
	Timeout     time.Duration // 传输超时时间
	DelayTime   time.Duration // 重试间隔时间
	Data        url.Values
	Retrytimes  int // 重试次数
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

func (req *RequestOption) Retry(n int, t time.Duration) *RequestOption {
	req.Retrytimes = n
	req.DelayTime = t
	return req
}

// 转码
func (r *Response) Charconv(c string) *Response {
	resp := r.Response
	dec := mahonia.NewDecoder(c)
	resp.Body = ioutil.NopCloser(dec.NewReader(resp.Body))
	return r
}

func (response *Response) Bytes() ([]byte, error) {
	defer response.Response.Body.Close()
	return ioutil.ReadAll(response.Response.Body)
}

func (response *Response) Json(t interface{}) error {
	defer response.Response.Body.Close()
	data, err := ioutil.ReadAll(response.Response.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(data, t)
	return nil
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
