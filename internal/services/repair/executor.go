package repair

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// SSHConfig SSH 连接配置
type SSHConfig struct {
	Host     string // IP 或域名
	Port     int    // 默认 22
	User     string
	Password string
	Timeout  time.Duration
}

// SSHExecutor 基于 SSH 的远程命令执行器
type SSHExecutor struct{}

// NewSSHExecutor 创建 SSH 执行器
func NewSSHExecutor() *SSHExecutor {
	return &SSHExecutor{}
}

// ExecResult 命令执行结果
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Execute 在目标主机上执行命令，返回结果
func (e *SSHExecutor) Execute(cfg SSHConfig, command, cwd string, timeout int) (*ExecResult, error) {
	port := cfg.Port
	if port == 0 {
		port = 22
	}
	deadline := time.Duration(timeout) * time.Second
	if deadline <= 0 {
		deadline = 30 * time.Second
	}
	if cfg.Timeout > 0 {
		deadline = cfg.Timeout
	}

	// 1. 建立 SSH 连接
	sshCfg := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应改为已知主机验证
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败 [%s]: %v", addr, err)
	}
	defer func() { _ = client.Close() }()

	// 2. 打开 Session
	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建 SSH Session 失败: %v", err)
	}
	defer func() { _ = session.Close() }()

	// 3. 捕获输出
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// 4. 拼接工作目录（如果有）
	fullCmd := command
	if cwd != "" && cwd != "/" {
		fullCmd = fmt.Sprintf("cd %s && %s", cwd, command)
	}

	// 5. 带超时执行
	start := time.Now()
	done := make(chan error, 1)
	go func() {
		done <- session.Run(fullCmd)
	}()

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(deadline):
		_ = session.Signal(ssh.SIGKILL)
		return nil, fmt.Errorf("命令执行超时（%v）: %s", deadline, command)
	}

	duration := time.Since(start)
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			exitCode = -1
		}
	}

	return &ExecResult{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// ExecuteWithStream 执行命令并逐行回调实时输出（用于 SSE 推送）
func (e *SSHExecutor) ExecuteWithStream(cfg SSHConfig, command, cwd string, timeout int, lineCallback func(line string)) (*ExecResult, error) {
	port := cfg.Port
	if port == 0 {
		port = 22
	}
	deadline := time.Duration(timeout) * time.Second
	if deadline <= 0 {
		deadline = 30 * time.Second
	}

	sshCfg := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(cfg.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", port))
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return nil, fmt.Errorf("SSH 连接失败 [%s]: %v", addr, err)
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("创建 SSH Session 失败: %v", err)
	}
	defer func() { _ = session.Close() }()

	// 使用管道实现逐行回调
	outPipe, err := session.StdoutPipe()
	if err != nil {
		return nil, err
	}
	errPipe, err := session.StderrPipe()
	if err != nil {
		return nil, err
	}

	fullCmd := command
	if cwd != "" && cwd != "/" {
		fullCmd = fmt.Sprintf("cd %s && %s", cwd, command)
	}

	start := time.Now()
	if err := session.Start(fullCmd); err != nil {
		return nil, fmt.Errorf("启动命令失败: %v", err)
	}

	var allOutput strings.Builder
	// 读取 stdout
	go func() {
		buf := make([]byte, 4096)
		var line strings.Builder
		for {
			n, err := outPipe.Read(buf)
			if n > 0 {
				chunk := string(buf[:n])
				allOutput.WriteString(chunk)
				for _, ch := range chunk {
					if ch == '\n' {
						ln := line.String()
						line.Reset()
						if lineCallback != nil {
							lineCallback(ln)
						}
					} else {
						line.WriteRune(ch)
					}
				}
			}
			if err != nil {
				break
			}
		}
	}()

	// 读取 stderr
	var errOutput strings.Builder
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := errPipe.Read(buf)
			if n > 0 {
				errOutput.WriteString(string(buf[:n]))
			}
			if err != nil {
				break
			}
		}
	}()

	done := make(chan error, 1)
	go func() { done <- session.Wait() }()

	var runErr error
	select {
	case runErr = <-done:
	case <-time.After(deadline):
		_ = session.Signal(ssh.SIGKILL)
		return nil, fmt.Errorf("命令执行超时（%v）", deadline)
	}

	duration := time.Since(start)
	exitCode := 0
	if runErr != nil {
		if exitErr, ok := runErr.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			exitCode = -1
		}
	}

	return &ExecResult{
		Stdout:   strings.TrimSpace(allOutput.String()),
		Stderr:   strings.TrimSpace(errOutput.String()),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}
