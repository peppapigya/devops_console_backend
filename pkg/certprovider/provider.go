package certprovider

type CertificateResult struct {
	Daemon            string // 需要监控的域名
	CertPem           []byte //公钥
	KeyPem            []byte // 私钥
	IssuerCertificate []byte // 中级证书
	ExpireTime        string // 过期时间
	Expired           bool   // 是否过期

}
type Provider interface {
	GetCertificateDetail(req GetCertificateDetailRequest) (*CertificateResult, error) // 获取证书信息

	SyncCertificate() error // 同步证书信息

	UploadCertificate(req UploadCertificateRequest) (int64, error) // 上传证书信息
	ApplyCertificate(instanceId string) error

	DeleteCertificate(certId int64) error // 删除证书信息
}

type GetCertificateDetailRequest struct {
	CertId     int64
	CertFilter bool
}
type UploadCertificateRequest struct {
	Name *string // 名字必传
	Crt  *string // 公钥
	Key  *string // 私钥
}
