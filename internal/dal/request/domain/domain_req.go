package domain

// ==================== 域名监控 ====================

type DomainPageReq struct {
	Page         int    `form:"page" binding:"required,min=1"`
	PageSize     int    `form:"pageSize" binding:"required,min=1,max=200"`
	Domain       string `form:"domain"`
	Status       string `form:"status"`
	AliveStatus  string `form:"aliveStatus"`
	Tag          string `form:"tag"`
	ExpireWithin int    `form:"expireWithin"` // 即将过期天数
}

type DomainCreateReq struct {
	Domain        string   `json:"domain" binding:"required"`
	Tags          []string `json:"tags"`
	Protocol      string   `json:"protocol"`
	CheckInterval int      `json:"checkInterval"`
	Enabled       *bool    `json:"enabled"`
	Remark        string   `json:"remark"`
}

type DomainUpdateReq struct {
	Tags          []string `json:"tags"`
	Protocol      string   `json:"protocol"`
	CheckInterval int      `json:"checkInterval"`
	Enabled       *bool    `json:"enabled"`
	Remark        string   `json:"remark"`
}

type DomainToggleReq struct {
	Enabled bool `json:"enabled"`
}

// ==================== SSL 证书 ====================

type SslCertPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Domain   string `form:"domain"`
	Status   *int8  `form:"status"`
}

type ApplySslCertReq struct {
	Domain       string  `json:"domain" binding:"required"`
	DnsConfigID  *uint64 `json:"dnsConfigId" binding:"required"`
	CertSource   string  `json:"certSource" binding:"required,oneof=ACME CAS"`
	CAProvider   string  `json:"caProvider" binding:"required"`
	KeyAlgorithm string  `json:"keyAlgorithm"`
	Email        string  `json:"email" binding:"required,email"`
}

type UploadSslCertReq struct {
	Domain  string `json:"domain" binding:"required"`
	CertPem string `json:"certPem" binding:"required"`
	KeyPem  string `json:"keyPem" binding:"required"`
}

// ==================== DNS 云厂商配置 ====================

type DnsProviderPageReq struct {
	Page     int    `form:"page" binding:"required,min=1"`
	PageSize int    `form:"pageSize" binding:"required,min=1,max=200"`
	Name     string `form:"name"`
	Status   string `form:"status"`
}

type DnsProviderCreateReq struct {
	Name         string `json:"name" binding:"required,min=1,max=200"`
	Provider     string `json:"provider" binding:"required"`
	AccessKey    string `json:"accessKey" binding:"required"`
	AccessSecret string `json:"accessSecret" binding:"required"`
	ZoneID       string `json:"zoneId"`
	Region       string `json:"region"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
}

type DnsProviderUpdateReq struct {
	Name         string `json:"name" binding:"required,min=1,max=200"`
	Provider     string `json:"provider" binding:"required"`
	AccessKey    string `json:"accessKey" binding:"required"`
	AccessSecret string `json:"accessSecret"`
	ZoneID       string `json:"zoneId"`
	Region       string `json:"region"`
	Email        string `json:"email" binding:"required,email"`
	Phone        string `json:"phone"`
	Status       string `json:"status"`
}
