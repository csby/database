package mssql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Connection struct {
	Server   string `json:"server"`   // 服务器名称或IP, 默认127.0.0.1
	Port     int    `json:"port"`     // 服务器端口, 默认3306
	Instance string `json:"instance"` // 数据库实例, 默认MSSQLSERVER
	Schema   string `json:"schema"`   // 数据库名称
	User     string `json:"user"`     // 登录名
	Password string `json:"password"` // 登陆密码
	Timeout  int    `json:"Timeout" note:"连接超时时间，单位秒，默认10"`
}

func (s *Connection) DriverName() string {
	return "sqlserver"
}

func (s *Connection) SourceName() string {
	// sqlserver://username:password@host/instance?param1=value&param2=value
	// sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30 // `SQLExpress instance
	q := url.Values{}
	if s.Timeout > 0 {
		q.Add("&connection+timeout=", fmt.Sprint(s.Timeout))
	}

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(s.User, s.Password),
		Host:   fmt.Sprintf("%s:%d", s.Server, s.Port),
	}

	if len(s.Instance) > 0 {
		if strings.ToUpper(s.Instance) != "MSSQLSERVER" {
			u.Path = s.Instance
		}
	}
	if len(q) > 0 {
		u.RawQuery = q.Encode()
	}

	return u.String()
}

func (s *Connection) SchemaName() string {
	return s.Schema
}

func (s *Connection) SaveToFile(filePath string) error {
	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	fileFolder := filepath.Dir(filePath)
	_, err = os.Stat(fileFolder)
	if os.IsNotExist(err) {
		os.MkdirAll(fileFolder, 0777)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, string(bytes[:]))

	return err
}

func (s *Connection) LoadFromFile(filePath string) error {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}
