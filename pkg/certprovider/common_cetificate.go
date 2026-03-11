package certprovider

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"
)

type CertificateStatus struct {
	Issuer     string    // 证书颁发商
	ExpireTime time.Time // 过期时间
	DaysLeft   int       // 剩余天数
	IsExpired  bool      // 状态
	StatueCode int
	Latency    int // 秒数
}

// CheckCertificateStatus 检查证书是否过期
func CheckCertificateStatus(name string) (*CertificateStatus, error) {
	address := name + ":443"
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	dialer := &net.Dialer{
		Timeout: 10 * time.Second,
	}
	conn, err := tls.DialWithDialer(dialer, "tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("TLS 连接失败：%v", err.Error())
	}
	defer func() { _ = conn.Close() }()
	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	startTime := time.Now()
	resp, err := httpClient.Get("https://" + name + "/")
	if err != nil {
		return nil, err
	}
	latency := time.Since(startTime)
	defer func() { _ = resp.Body.Close() }()
	cert := conn.ConnectionState().PeerCertificates[0]
	daysLeft := int(time.Until(cert.NotAfter).Hours() / 24)
	now := time.Now()
	return &CertificateStatus{
		Issuer:     cert.Issuer.CommonName,
		ExpireTime: cert.NotAfter,
		DaysLeft:   daysLeft,
		IsExpired:  now.After(cert.NotAfter),
		StatueCode: resp.StatusCode,
		Latency:    int(latency / time.Millisecond),
	}, nil

}
