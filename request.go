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
	"net/url"
	"strings"
	"time"
)

type RequestOption struct {
	Method       string
	Url          string
	Header       http.Header
	ConnTimeout  time.Duration     // 链接超时时间
	Timeout      time.Duration     // 传输超时时间
	Retrytimes   int               // 重试次数
	DelayTime    time.Duration     // 重试间隔时间
	ParamsData   map[string]string // GET参数
	FormData     url.Values        // post data
	JsonFormData *bytes.Reader     // json post data
	JsonContent  bool              // Content-Type是否为json
}

type Response struct {
	Response *http.Response
	Error    error
}

func Request(url string, method string) *RequestOption {
	// 默认参数
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

// get参数
func (option *RequestOption) Params(params map[string]string) *RequestOption {
	option.ParamsData = params
	return option
}

// post form data
func (option *RequestOption) Form(data url.Values) *RequestOption {
	option.FormData = data
	return option
}

// json post form data
func (option *RequestOption) JsonForm(data *map[string]interface{}) *RequestOption {
	option.JsonContent = true
	bytesData, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		return option
	}
	option.JsonFormData = bytes.NewReader(bytesData)
	return option
}

// 请求
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

	// request
	var req *http.Request
	var reqerr error
	if option.JsonContent {
		req, reqerr = http.NewRequest(option.Method, option.Url, option.JsonFormData) // Content-Type 为json时
	} else {
		req, reqerr = http.NewRequest(option.Method, option.Url, option.FormData)
	}
	if reqerr != nil {
		result.Error = reqerr
		return result
	}

	// get 参数
	q := req.URL.Query()
	for k, v := range option.ParamsData {
		q.Add(k, v)
	}

	// 默认headers
	if option.JsonContent {
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
	} else {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// 请求失败重新请求
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

func (response *Response) Native() (*http.Response, error) {
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
