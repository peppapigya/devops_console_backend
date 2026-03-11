package test

import (
	"devops-console-backend/pkg/certprovider"
	"testing"
)

func TestCheckCertificateStatus(t *testing.T) {
	domain := "www.baidu.com"

	status, err := certprovider.CheckCertificateStatus(domain)
	if err != nil {
		t.Fatalf("查询失败: %v", err)
	}

	t.Logf("--- 证书巡检报告 ---")
	t.Logf("域名: %s", domain)
	t.Logf("颁发商: %s", status.Issuer)
	t.Logf("过期时间: %v", status.ExpireTime.Local())
	t.Logf("剩余天数: %d 天", status.DaysLeft)

	if status.DaysLeft < 7 {
		t.Errorf("警告：证书即将过期，请尽快更换！")
	}
}
