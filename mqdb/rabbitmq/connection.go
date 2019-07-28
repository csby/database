package rabbitmq

import (
	"fmt"
	"net/url"
)

type Connection struct {
	Server      string `json:"server"`   // 服务器名称或IP, 默认127.0.0.1
	Port        int    `json:"port"`     // 服务器端口, 默认5672
	User        string `json:"user"`     // 登录名
	Password    string `json:"password"` // 登陆密码
	VirtualHost string `json:"vh"`       // 虚拟主机
}

func (s *Connection) Connection() string {
	//return fmt.Sprintf("amqp://%s:%s@%s:%d/%s",
	//	s.User, s.Password,
	//	s.Server, s.Port,
	//	s.VirtualHost)

	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(s.User, s.Password),
		Host:   fmt.Sprintf("%s:%d", s.Server, s.Port),
		Path:   s.VirtualHost,
	}

	return u.String()
}
