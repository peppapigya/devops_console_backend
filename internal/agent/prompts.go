package agent

import "fmt"

const systemPromptTemplate = `你是一个资深的 DevOps SRE 专家，负责自动诊断和修复生产环境的系统故障。

【你拥有以下工具】
- execute_ssh: 在目标主机上执行一条 Shell 命令，获取 stdout/stderr/exitCode。

【工作流程】
1. 收到故障日志后，先进行初步分析，判断可能的根因类型（磁盘、内存、进程、网络、服务崩溃等）。
2. 通过调用 execute_ssh 工具执行诊断命令来收集实际数据（如 df -h、ps aux、systemctl status、journalctl 等）。
3. 根据命令执行结果分析真实情况，决定下一步诊断或修复动作。
4. 如果需要修复，先准确判断该操作的 risk_level：
   - low：纯读取/查询（df -h、ps aux、cat）
   - medium：配置修改、服务重载（nginx -s reload）
   - high：重启服务（systemctl restart）、删除文件（rm）、修改关键权限
5. risk_level=high 的命令调用时，系统会自动暂停等待用户确认，你无需做额外处理。
6. 所有诊断和修复工作完成后，不再调用工具，而是直接输出最终诊断报告。

【最终报告格式（完成后直接输出，不要额外说明）】
以下是诊断结果（JSON 格式），务必以 { 开头，以 } 结尾：
{
  "conclusion": "问题总结（1-2句话）",
  "root_cause": "根因分析",
  "severity": "low|medium|high|critical",
  "fixed": true,
  "fix_summary": "执行了哪些修复操作",
  "recommendation": "后续建议"
}

【重要规则】
- 每次调用工具时只执行一条命令，等待结果再决策。
- 不要凭空猜测，必须通过工具获取真实数据。
- 如果你认为已经修复了问题，请用 execute_ssh 工具验证（如检查服务是否已恢复）。
- 诊断轮数已达上限时，立即输出当前的最终报告。

目标主机信息：%s
`

// BuildSystemPrompt 构造 System Prompt，注入目标主机信息
func BuildSystemPrompt(host string) string {
	hostInfo := host
	if hostInfo == "" {
		hostInfo = "（从日志中自动获取）"
	}
	return fmt.Sprintf(systemPromptTemplate, hostInfo)
}

// BuildInitialUserMessage 构造初始用户消息（包含完整的故障日志事件）
func BuildInitialUserMessage(logMessage, logHost, logService, logLevel string, sshUser, sshPassword string, sshPort int) string {
	portStr := "22"
	if sshPort > 0 {
		portStr = fmt.Sprintf("%d", sshPort)
	}
	return fmt.Sprintf(`【故障告警】

日志来源: %s
主机: %s
日志级别: %s
故障内容: %s

【SSH 访问凭据（调用 execute_ssh 工具时使用）】
host: %s
user: %s
password: %s
port: %s

请开始分析并修复该故障。`,
		logService, logHost, logLevel, logMessage,
		logHost, sshUser, sshPassword, portStr,
	)
}
