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
	Host     string `json:"host" note:"服务器名称或IP, 默认127.0.0.1"`
	Port     int    `json:"port" note:"服务器端口, 默认3306"`
	Instance string `json:"instance" note:"数据库实例, 默认MSSQLSERVER"`
	Schema   string `json:"schema" note:"数据库名称"`
	User     string `json:"user" note:"登录名"`
	Password string `json:"password" note:"登陆密码"`
	Timeout  int    `json:"timeout" note:"连接超时时间，单位秒，默认10"`
}

func (s *Connection) DriverName() string {
	return "sqlserver"
}

func (s *Connection) SourceName() string {
	// sqlserver://username:password@host/instance?param1=value&param2=value
	// sqlserver://sa@localhost/SQLExpress?database=master&connection+timeout=30 // `SQLExpress instance
	q := url.Values{}
	if len(s.Schema) > 0 {
		q.Add("database", s.Schema)
	}
	if s.Timeout > 0 {
		q.Add("&connection+timeout", fmt.Sprint(s.Timeout))
	}

	u := &url.URL{
		Scheme: "sqlserver",
		User:   url.UserPassword(s.User, s.Password),
		Host:   fmt.Sprintf("%s:%d", s.Host, s.Port),
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

func (s *Connection) CopyTo(target *Connection) int {
	if target == nil {
		return 0
	}

	count := 0
	if target.Host != s.Host {
		target.Host = s.Host
		count++
	}
	if target.Port != s.Port {
		target.Port = s.Port
		count++
	}
	if target.Instance != s.Instance {
		target.Instance = s.Instance
		count++
	}
	if target.Schema != s.Schema {
		target.Schema = s.Schema
		count++
	}

	if target.User != s.User {
		target.User = s.User
		count++
	}
	if target.Password != s.Password {
		target.Password = s.Password
		count++
	}
	if target.Timeout != s.Timeout {
		target.Timeout = s.Timeout
		count++
	}

	return count
}
