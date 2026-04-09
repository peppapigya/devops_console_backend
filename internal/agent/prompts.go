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
6. 一旦你通过命令定位了明确的根本错误原因（如确认了代码缺失分号等）：
   - 如果属于易于修复的常见配置错误，立刻使用 sed 或其他合适命令直接在目标机器上修改修复！
   - 修复后务必执行验证命令（如 nginx -t）确保解决。
7. 【最核心步骤】一切排查或修复完成后，你必须唯一调用 'submit_diagnosis_report' 工具来结束流程并提交诊断报告！！！绝不能再输出常规文本！！！

【致命红线限制】
- 每次调用工具时只能执行一条简单命令。
- 如果排查陷入死胡同或你已完成文件的修复并验证成功，必须并只能立刻调用 'submit_diagnosis_report' 工具结束整个分析！
- 严禁为了确认细枝末节而陷入 ls/cat/head等周边检查的无限死循环。只要错误查明并修正，立刻停止！

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
