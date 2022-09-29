package oracle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
)

const (
	connectionStringFormat = "oracle://%s@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=tcp)(HOST=%s)(PORT=%d)))(CONNECT_DATA=(SERVICE_NAME=%s)))"
)

type Connection struct {
	Host     string   `json:"host" note:"服务器名称或IP, 默认127.0.0.1"`
	Port     int      `json:"port" note:"服务器端口, 默认1521"`
	SID      string   `json:"sid" note:"SID"`
	User     string   `json:"user" note:"登录名"`
	Password string   `json:"password" note:"登陆密码"`
	Owners   []string `json:"owners" note"所有者，用于生成表结构"`
}

func (s *Connection) DriverName() string {
	return "goracle"
}

func (s *Connection) SourceName() string {
	// user/pass@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=tcp)(HOST=hostname)(PORT=port)))(CONNECT_DATA=(SERVICE_NAME=sn)))
	// oracle://user:pass@(DESCRIPTION=(ADDRESS_LIST=(ADDRESS=(PROTOCOL=tcp)(HOST=hostname)(PORT=port)))(CONNECT_DATA=(SERVICE_NAME=sn)))

	return fmt.Sprintf(connectionStringFormat,
		url.UserPassword(s.User, s.Password).String(),
		s.Host,
		s.Port,
		s.SID)

}

func (s *Connection) ClusterSourceName(readOnly bool) string {
	return s.SourceName()
}

func (s *Connection) SchemaName() string {
	return s.SID
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
	if target.SID != s.SID {
		target.SID = s.SID
		count++
	}
	if target.SID != s.SID {
		target.SID = s.SID
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

	owners, changeCount := s.copy(s.Owners, target.Owners)
	if changeCount > 0 {
		target.Owners = owners
		count++
	}

	return count
}

func (s *Connection) copy(source, target []string) ([]string, int) {
	results := make([]string, 0)
	resultCount := 0

	sourceCount := len(source)
	targetCount := len(target)
	if sourceCount == targetCount {
		for index := 0; index < sourceCount; index++ {
			sourceValue := source[index]
			results = append(results, sourceValue)
			if sourceValue != target[index] {
				resultCount++
			}
		}
	} else {
		for index := 0; index < sourceCount; index++ {
			sourceValue := source[index]
			results = append(results, sourceValue)
			resultCount++
		}
	}

	return results, resultCount
}
