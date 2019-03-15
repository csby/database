package mysql

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Connection struct {
	Server   string `json:"server" note:"服务器名称或IP, 默认127.0.0.1"`
	Port     int    `json:"port" note:"服务器端口, 默认3306"`
	Schema   string `json:"schema" note:"数据库名称, 默认mysql"`
	Charset  string `json:"charset" note:"字符集, 默认utf8"`
	Timeout  int    `json:"timeout" note:"连接超时时间，单位秒，默认10"`
	User     string `json:"user" note:"登录名"`
	Password string `json:"password" note:"登陆密码"`
}

func (s *Connection) DriverName() string {
	return "mysql"
}

func (s *Connection) SourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&timeout=%ds&parseTime=true&loc=Local",
		s.User,
		s.Password,
		s.Server,
		s.Port,
		s.Schema,
		s.Charset,
		s.Timeout)
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
