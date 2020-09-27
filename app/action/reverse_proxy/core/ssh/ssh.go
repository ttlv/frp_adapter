package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"time"
)

func NewSshClient(FrpsPublicIpAddress, port string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout: time.Second * 5,
		User:    "root",
		//TODO 暂时先用明文密码登录,等基础逻辑全部实现之后再从ras key登录
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //这个可以， 但是不够安全
		//HostKeyCallback: hostKeyCallBackFunc(h.Host),
	}
	//if h.Type == "password" {
	config.Auth = []ssh.AuthMethod{ssh.Password("123456")}
	//} else {
	//	config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(h.Key)}
	//}
	FrpsPublicIpAddress = "10.1.11.38"
	c, err := ssh.Dial("tcp", fmt.Sprintf("%v:%v", FrpsPublicIpAddress, port), config)
	if err != nil {
		return nil, err
	}
	return c, nil
}
