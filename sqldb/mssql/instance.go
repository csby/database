package mssql

import (
	"fmt"
	"net"
	"strings"
)

type Instance struct {
	name    string
	port    string
	version string
}

func (s *Instance) Name() string {
	return s.name
}

func (s *Instance) Port() string {
	return s.port
}

func (s *Instance) Version() string {
	return s.version
}

func parseInstances(msg []byte) map[string]map[string]string {
	results := map[string]map[string]string{}
	if len(msg) > 3 && msg[0] == 5 {
		out_s := string(msg[3:])
		tokens := strings.Split(out_s, ";")
		instdict := map[string]string{}
		got_name := false
		var name string
		for _, token := range tokens {
			if got_name {
				instdict[name] = token
				got_name = false
			} else {
				name = token
				if len(name) == 0 {
					if len(instdict) == 0 {
						break
					}
					results[strings.ToUpper(instdict["InstanceName"])] = instdict
					instdict = map[string]string{}
					continue
				}
				got_name = true
			}
		}
	}
	return results
}

func getInstances(host, port string) (map[string]map[string]string, error) {

	conn, err := net.Dial("udp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	_, err = conn.Write([]byte{3})
	if err != nil {
		return nil, err
	}

	var resp = make([]byte, 16*1024-1)
	read, err := conn.Read(resp)
	if err != nil {
		return nil, err
	}

	return parseInstances(resp[:read]), nil
}
