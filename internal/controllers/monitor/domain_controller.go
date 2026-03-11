package monitor

import (
	"devops-console-backend/internal/common"
	"devops-console-backend/internal/dal/mapper"
	"devops-console-backend/internal/dal/model"
	reqDomain "devops-console-backend/internal/dal/request/domain"
	"devops-console-backend/pkg/certprovider"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DomainController 域名管理控制器
type DomainController struct {
	domainMapper      *mapper.DomainMapper
	sslCertMapper     *mapper.SslCertMapper
	dnsProviderMapper *mapper.DnsProviderMapper
}

func NewDomainController(dm *mapper.DomainMapper, sm *mapper.SslCertMapper, pm *mapper.DnsProviderMapper) *DomainController {
	return &DomainController{domainMapper: dm, sslCertMapper: sm, dnsProviderMapper: pm}
}

// ==================== 域名统计 ====================

func (c *DomainController) Stats(ctx *gin.Context) {
	stats, err := c.domainMapper.Stats()
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": stats})
}

// ==================== 域名 CRUD ====================

func (c *DomainController) ListDomains(ctx *gin.Context) {
	var req reqDomain.DomainPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.domainMapper.ListPage(req.Page, req.PageSize, req.Domain, req.Status, req.ExpireWithin)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

func (c *DomainController) CreateDomain(ctx *gin.Context) {
	var req reqDomain.DomainCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	remark := req.Remark
	interval := req.CheckInterval
	if interval == 0 {
		interval = 300
	}
	proto := req.Protocol
	if proto == "" {
		proto = "https"
	}
	d := &model.MonitorDomain{
		Domain:        req.Domain,
		Tags:          req.Tags,
		Protocol:      proto,
		CheckInterval: interval,
		Enabled:       enabled,
		Status:        "unknown",
		Remark:        &remark,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if enabled == true {
		status, err := certprovider.CheckCertificateStatus(req.Domain)
		if err == nil {
			d.CertProvider = &status.Issuer
			d.SslDaysLeft = &status.DaysLeft
			d.Status = "normal"
			if status.IsExpired {
				d.Status = "abnormal"
			}
			d.StatusCode = &status.StatueCode
			d.ResponseTime = &status.Latency
		}
	}

	if err := c.domainMapper.Create(d); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": d.ID}})
}

func (c *DomainController) UpdateDomain(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqDomain.DomainUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	d, err := c.domainMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "域名不存在")
		return
	}
	now := time.Now()
	d.Tags = req.Tags
	if req.Protocol != "" {
		d.Protocol = req.Protocol
	}
	if req.CheckInterval > 0 {
		d.CheckInterval = req.CheckInterval
	}
	if req.Enabled != nil {
		d.Enabled = *req.Enabled
	}
	d.Remark = &req.Remark
	d.UpdatedAt = &now
	if err := c.domainMapper.Update(d); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func (c *DomainController) DeleteDomain(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.domainMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func (c *DomainController) ToggleDomain(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqDomain.DomainToggleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	if err := c.domainMapper.UpdateFields(id, map[string]interface{}{"enabled": req.Enabled}); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

// ==================== SSL 证书 ====================

func (c *DomainController) ListSslCerts(ctx *gin.Context) {
	var req reqDomain.SslCertPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.sslCertMapper.ListPage(req.Page, req.PageSize, req.Domain, req.Status)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

func (c *DomainController) ApplySslCert(ctx *gin.Context) {
	var req reqDomain.ApplySslCertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	// 查找 DNS 配置名称
	dnsConfig := ""
	if req.DnsConfigID != nil {
		if p, err := c.dnsProviderMapper.GetByID(*req.DnsConfigID); err == nil {
			dnsConfig = p.Name
		}
	}
	now := time.Now()
	algo := req.KeyAlgorithm
	if algo == "" {
		algo = "EC256"
	}
	cert := &model.MonitorSslCert{
		Domain:       req.Domain,
		DnsConfigID:  req.DnsConfigID,
		DnsConfig:    dnsConfig,
		CertSource:   req.CertSource,
		CAProvider:   req.CAProvider,
		KeyAlgorithm: algo,
		Email:        req.Email,
		Status:       -1, // 申请中
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}

	// TODO: 根据 req.CertSource 和 req.CAProvider 调用对应的 SDK (如阿里云 ssl证书服务、腾讯云相关服务)
	// 例如：调用 Aliyun SDK (CreateCertificateRequest)
	// resp, err := aliyunSslClient.CreateCertificate(...)
	// 拿到第三方返回的唯一标识 (certId)，存入数据库：
	// dummyCertId := "cas-" + time.Now().Format("20060102150405")
	// cert.CertID = &dummyCertId

	if err := c.sslCertMapper.Create(cert); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": cert.ID, "message": "证书申请已提交，请稍后查看状态"}})
}

func (c *DomainController) UploadSslCert(ctx *gin.Context) {
	var req reqDomain.UploadSslCertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	cert := &model.MonitorSslCert{
		Domain:     req.Domain,
		CertSource: "上传",
		CAProvider: "手动",
		Status:     1, // 已签发
		CertPem:    &req.CertPem,
		KeyPem:     &req.KeyPem,
		CreatedAt:  &now,
		UpdatedAt:  &now,
	}
	if err := c.sslCertMapper.Create(cert); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": cert.ID}})
}

func (c *DomainController) DeleteSslCert(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.sslCertMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func (c *DomainController) DownloadSslCert(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	cert, err := c.sslCertMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "证书不存在")
		return
	}
	if cert.CertPem == nil {
		common.FailWithMsg(ctx, "证书内容不存在")
		return
	}
	ctx.Header("Content-Disposition", "attachment; filename="+cert.Domain+".crt")
	ctx.Header("Content-Type", "application/x-pem-file")
	ctx.String(200, *cert.CertPem)
}

// ==================== DNS 云厂商配置 ====================

func (c *DomainController) ListDnsProviders(ctx *gin.Context) {
	var req reqDomain.DnsProviderPageReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	total, list, err := c.dnsProviderMapper.ListPage(req.Page, req.PageSize, req.Name, req.Status)
	if err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"total": total, "list": list}})
}

func (c *DomainController) CreateDnsProvider(ctx *gin.Context) {
	var req reqDomain.DnsProviderCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	now := time.Now()
	status := req.Status
	if status == "" {
		status = "active"
	}
	p := &model.MonitorDnsProvider{
		Name:         req.Name,
		Provider:     req.Provider,
		AccessKey:    req.AccessKey,
		AccessSecret: req.AccessSecret,
		Email:        req.Email,
		Phone:        req.Phone,
		Status:       status,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	if req.ZoneID != "" {
		p.ZoneID = &req.ZoneID
	}
	if req.Region != "" {
		p.Region = &req.Region
	}
	if err := c.dnsProviderMapper.Create(p); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": gin.H{"id": p.ID}})
}

func (c *DomainController) UpdateDnsProvider(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	var req reqDomain.DnsProviderUpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}
	p, err := c.dnsProviderMapper.GetByID(id)
	if err != nil {
		common.FailWithMsg(ctx, "配置不存在")
		return
	}
	now := time.Now()
	p.Name = req.Name
	p.Provider = req.Provider
	p.AccessKey = req.AccessKey
	if req.AccessSecret != "" {
		p.AccessSecret = req.AccessSecret
	}
	p.Email = req.Email
	p.Phone = req.Phone
	if req.Status != "" {
		p.Status = req.Status
	}
	if req.ZoneID != "" {
		p.ZoneID = &req.ZoneID
	}
	if req.Region != "" {
		p.Region = &req.Region
	}
	p.UpdatedAt = &now
	if err := c.dnsProviderMapper.Update(p); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func (c *DomainController) DeleteDnsProvider(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if err := c.dnsProviderMapper.SoftDelete(id); err != nil {
		common.FailWithError(ctx, err)
		return
	}
	common.Success(ctx, gin.H{"data": nil})
}

func (c *DomainController) TestDnsProvider(ctx *gin.Context) {
	id, err := parseDomainUint64Param(ctx, "id")
	if err != nil {
		common.Fail(ctx, common.BadRequest)
		return
	}
	if _, err := c.dnsProviderMapper.GetByID(id); err != nil {
		common.FailWithMsg(ctx, "配置不存在")
		return
	}
	// TODO: 实际调用各云厂商 DNS API 做连通性测试
	common.Success(ctx, gin.H{"data": gin.H{"message": "连接测试成功（需实现各厂商 SDK 调用）"}})
}

func parseDomainUint64Param(ctx *gin.Context, key string) (uint64, error) {
	return strconv.ParseUint(ctx.Param(key), 10, 64)
}
