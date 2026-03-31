package executor

import (
	"bytes"
	"context"
	utils "devops-console-backend/pkg/utils/aes"
	"fmt"
	"log"
	"os/exec"
	"time"

	"devops-console-backend/internal/dal/model"
	"devops-console-backend/pkg/configs"

	"golang.org/x/crypto/ssh"
)

// ScriptExecutor 脚本执行器
type ScriptExecutor struct{}

func NewScriptExecutor() *ScriptExecutor {
	return &ScriptExecutor{}
}

func (e *ScriptExecutor) GetType() string {
	return "script"
}

func (e *ScriptExecutor) Validate(config map[string]interface{}) error {
	content := getString(config, "content", "")
	if content == "" {
		return fmt.Errorf("脚本内容不能为空")
	}
	return nil
}

func (e *ScriptExecutor) Execute(ctx context.Context, execCtx *TaskExecutionContext) *ExecutionResult {
	startTime := time.Now()

	scriptType := getString(execCtx.Config, "script_type", "shell")
	content := getString(execCtx.Config, "content", "")
	workingDir := getString(execCtx.Config, "working_dir", "")

	if execCtx.TargetID > 0 && execCtx.TargetType == "host" {
		return e.executeRemote(ctx, execCtx, scriptType, content, workingDir, startTime)
	}

	return e.executeLocal(ctx, execCtx, scriptType, content, workingDir, startTime)
}

func (e *ScriptExecutor) executeLocal(ctx context.Context, execCtx *TaskExecutionContext, scriptType, content, workingDir string, startTime time.Time) *ExecutionResult {
	execCtx.Logger.Log("info", fmt.Sprintf("开始执行本地%s脚本", scriptType))
	if workingDir != "" {
		execCtx.Logger.Log("info", fmt.Sprintf("工作目录: %s", workingDir))
	}
	execCtx.Logger.Log("info", fmt.Sprintf("执行内容: \n%s", content))

	var cmd *exec.Cmd
	switch scriptType {
	case "shell":
		cmd = exec.CommandContext(ctx, "sh", "-c", content)
	case "python":
		cmd = exec.CommandContext(ctx, "python3", "-c", content)
	case "powershell":
		cmd = exec.CommandContext(ctx, "powershell", "-Command", content)
	default:
		cmd = exec.CommandContext(ctx, "sh", "-c", content)
	}

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime).Milliseconds()

	output := stdout.String()
	errorOutput := stderr.String()

	if output != "" {
		execCtx.Logger.Log("info", "脚本标准输出: \n"+output)
	}

	if errorOutput != "" {
		execCtx.Logger.Log("warn", "脚本错误输出: \n"+errorOutput)
	}

	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("脚本执行失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			Output:   output,
			ErrorMsg: fmt.Sprintf("%v\n%s", err, errorOutput),
			Duration: duration,
		}
	}

	execCtx.Logger.Log("info", fmt.Sprintf("脚本执行成功，耗时: %dms", duration))

	return &ExecutionResult{
		Success:    true,
		Output:     output,
		Duration:   duration,
		OutputVars: map[string]interface{}{},
	}
}

func (e *ScriptExecutor) executeRemote(ctx context.Context, execCtx *TaskExecutionContext, scriptType, content, workingDir string, startTime time.Time) *ExecutionResult {
	var host model.AssetHost
	err := configs.GORMDB.Where("id = ? AND deleted_at IS NULL", execCtx.TargetID).First(&host).Error
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("获取主机信息失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("获取主机信息失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}

	execCtx.Logger.Log("info", fmt.Sprintf("开始远程执行%s脚本, 目标主机: %s:%d (ID: %d)", scriptType, host.IP, host.Port, host.ID))
	if workingDir != "" {
		execCtx.Logger.Log("info", fmt.Sprintf("工作目录: %s", workingDir))
	}
	execCtx.Logger.Log("info", fmt.Sprintf("执行内容: \n%s", content))

	var auth []ssh.AuthMethod
	if host.AuthType == "password" && host.Password != nil {
		log.Printf("加密的密码：%v", *host.Password)
		decryptedPwd := DecryptPassword(*host.Password)
		log.Printf("解析出来的密码：%v", decryptedPwd)
		auth = append(auth, ssh.Password(decryptedPwd))
	} else if host.AuthType == "key" && host.PrivateKey != nil {
		signer, err := ssh.ParsePrivateKey([]byte(*host.PrivateKey))
		if err != nil {
			execCtx.Logger.Log("error", fmt.Sprintf("解析私钥失败: %v", err))
			return &ExecutionResult{
				Success:  false,
				ErrorMsg: fmt.Sprintf("解析私钥失败: %v", err),
				Duration: time.Since(startTime).Milliseconds(),
			}
		}
		auth = append(auth, ssh.PublicKeys(signer))
	} else {
		if host.Password != nil {
			decryptedPwd := DecryptPassword(*host.Password)
			auth = append(auth, ssh.Password(decryptedPwd))
		} else {
			execCtx.Logger.Log("error", "主机认证信息不完整或不支持的认证方式")
			return &ExecutionResult{
				Success:  false,
				ErrorMsg: "主机认证信息不完整或不支持的认证方式",
				Duration: time.Since(startTime).Milliseconds(),
			}
		}
	}

	config := &ssh.ClientConfig{
		User:            host.Username,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	execCtx.Logger.Log("info", fmt.Sprintf("正在连接主机 %s:%d...", host.IP, host.Port))
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", host.IP, host.Port), config)
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("SSH连接失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("SSH连接失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	defer func() { _ = client.Close() }()

	session, err := client.NewSession()
	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("创建SSH会话失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			ErrorMsg: fmt.Sprintf("创建SSH会话失败: %v", err),
			Duration: time.Since(startTime).Milliseconds(),
		}
	}
	defer func() { _ = session.Close() }()

	var finalCmd string
	switch scriptType {
	case "shell":
		finalCmd = content
	case "python":
		finalCmd = fmt.Sprintf("python3 << 'EOF'\n%s\nEOF", content)
	default:
		finalCmd = content
	}

	if workingDir != "" {
		finalCmd = fmt.Sprintf("cd %s && %s", workingDir, finalCmd)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(finalCmd)
	duration := time.Since(startTime).Milliseconds()

	output := stdout.String()
	errorOutput := stderr.String()

	if output != "" {
		execCtx.Logger.Log("info", "远程脚本标准输出: \n"+output)
	}

	if errorOutput != "" {
		execCtx.Logger.Log("warn", "远程脚本错误输出: \n"+errorOutput)
	}

	if err != nil {
		execCtx.Logger.Log("error", fmt.Sprintf("远程脚本执行失败: %v", err))
		return &ExecutionResult{
			Success:  false,
			Output:   output,
			ErrorMsg: fmt.Sprintf("%v\n%s", err, errorOutput),
			Duration: duration,
		}
	}

	execCtx.Logger.Log("info", fmt.Sprintf("远程脚本执行成功，耗时: %dms", duration))

	return &ExecutionResult{
		Success:    true,
		Output:     output,
		Duration:   duration,
		OutputVars: map[string]interface{}{},
	}
}

func DecryptPassword(encryptedPassword string) string {
	if encryptedPassword == "" {
		return ""
	}
	key, err := configs.GetEncryptionKey()
	if err != nil || key == nil {
		log.Printf("获取加密密钥失败: %v", err)
		log.Printf("使用原始密码,%s", key)
		return encryptedPassword
	}
	decrypted, err := utils.AESDecrypt(key, encryptedPassword)
	if err != nil {
		return encryptedPassword
	}
	return decrypted
}
