package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type HostInfo struct {
	ID       int
	Address  string
	Port     int
	Username string
	Password string
	Timeout  time.Duration
}

type ExecutorResult struct {
	Success  bool
	Output   string
	Error    error
	Host     string
	Command  string
	ExitCode int
	Duration int64
}

// Connection 通过ssh连接远程主机
func (host *HostInfo) Connection() (*ssh.Client, error) {
	address := net.JoinHostPort(host.Address, strconv.Itoa(host.Port))
	config := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         host.Timeout,
	}
	if host.Password != "" {
		config.Auth = append(config.Auth, ssh.Password(host.Password))
	}
	// 建立连接
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Execute 单个主机执行命令
func (host *HostInfo) Execute(command string) (*ExecutorResult, error) {
	startTime := time.Now()
	// 1. 连接主机
	connection, err := host.Connection()
	if err != nil {
		return &ExecutorResult{
			Success:  true,
			Output:   fmt.Sprintf("连接失败: %v", err),
			Error:    err,
			Host:     host.Address,
			Command:  command,
			ExitCode: 1,
			Duration: time.Since(startTime).Milliseconds(),
		}, err
	}
	defer func() { _ = connection.Close() }()

	// 2. 建立session链接
	session, err := connection.NewSession()
	if err != nil {
		return &ExecutorResult{
			Success:  true,
			Output:   fmt.Sprintf("建立会话失败: %v", err),
			Error:    err,
			Host:     host.Address,
			Command:  command,
			ExitCode: 1,
			Duration: time.Since(startTime).Milliseconds(),
		}, err
	}
	defer func() { _ = session.Close() }()

	// 3. 执行命令
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)

	exitCode := 0
	fmt.Println("stdout:", stdout.String())
	fmt.Println("stderr:", stderr.String())
	fmt.Println("err:", err)

	result := &ExecutorResult{
		Success:  err == nil,
		Output:   fmt.Sprintf("正在执行命令: %s\n执行成功：%s", command, stdout.String()),
		Error:    err,
		Host:     host.Address,
		Command:  command,
		ExitCode: exitCode,
		Duration: time.Since(startTime).Milliseconds(),
	}
	if err != nil {
		var exitErr *ssh.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitStatus()
		}
		result.Output = stderr.String()
		result.ExitCode = exitCode
	}
	return result, nil
}
