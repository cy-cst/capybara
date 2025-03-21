package capybara

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
)

// **** context
type Context interface {
	JSON(code int, data interface{}) error
	XML(code int, data interface{}) error
	String(code int, s string) error
	HTML(code int, html string) error

	Request() *http.Request
	GetHeader(key string) string
	Set(key string, value interface{})
	Get(key string) interface{}
	Bind(data interface{}) (err error)

	Param(name string) string
}

type context struct {
	w      http.ResponseWriter
	r      *http.Request
	data   map[string]interface{}
	capa   *capybara
	params map[string]string
}

// 应用到当前的 context
func (c *context) ApplyContext(cap *capybara, params map[string]string, w http.ResponseWriter, r *http.Request) {
	c.capa = cap
	c.w = w
	c.r = r
	c.data = make(map[string]interface{})
	c.params = params
}

func (c *context) JSON(code int, data interface{}) error {
	jsonEncoder := json.NewEncoder(c.w)
	err := jsonEncoder.Encode(data)
	if err != nil {
		return c.String(http.StatusInternalServerError, "解析json出错")
	}
	return nil
}

func (c *context) String(code int, s string) error {
	c.w.Header().Set("Content-Type", "text/plain")
	c.w.WriteHeader(code)
	_, err := c.w.Write([]byte(s))
	return err
}

func (c *context) XML(code int, data interface{}) error {
	c.w.Header().Set("Content-Type", "application/xml")
	c.w.WriteHeader(code)
	xmlEncoder := xml.NewEncoder(c.w)
	err := xmlEncoder.Encode(data)
	if err != nil {
		c.String(http.StatusInternalServerError, "解析xml出错")
	}
	return nil
}

func (c *context) HTML(code int, html string) error {
	c.w.Header().Set("Content-Type", "text/html")
	c.w.WriteHeader(code)
	_, err := c.w.Write([]byte(html))
	return err
}

func (c *context) Param(name string) string {
	// 需求：http://localhost:8080/user/123/post/456
	//                           /user/:id/post/:post_id
	// id ： 123
	// post_id : 456
	if _, ok := c.params[name]; !ok {
		return ""
	}
	return c.params[name]
}

func (c *context) Request() *http.Request {
	return c.r
}

func (c *context) GetHeader(key string) string {
	return c.r.Header.Get(key)
}

func (c *context) Get(key string) interface{} {
	return c.data[key]
}

func (c *context) Set(key string, value interface{}) {
	c.data[key] = value
}

func (c *context) Bind(data interface{}) (err error) {
	// 读取请求体
	body, err := io.ReadAll(c.r.Body)
	if err != nil {
		return err
	}
	defer c.r.Body.Close()

	// 如果请求体为空，返回错误
	if len(body) == 0 {
		return errors.New("request body is empty")
	}

	err = json.Unmarshal(body, data)
	if err != nil {
		return err
	}
	return err
}
