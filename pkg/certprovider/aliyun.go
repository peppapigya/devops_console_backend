package certprovider

import (
	"fmt"

	cas "github.com/alibabacloud-go/cas-20200407/v2/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	console "github.com/alibabacloud-go/tea-console/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type AliProvider struct {
	AccessKey string
	SecretKey string
	Email     string
}

func (p *AliProvider) GetCertificateDetail(req GetCertificateDetailRequest) (*CertificateResult, error) {
	client, err := CreateClient(p)
	if err != nil {
		return nil, err
	}
	request := &cas.GetUserCertificateDetailRequest{
		CertId: tea.Int64(req.CertId),
	}
	result, err := client.GetUserCertificateDetail(request)
	if err != nil {
		return nil, err
	}
	console.Log(util.ToJSONString(tea.ToMap(result)))
	return &CertificateResult{
		CertPem:           []byte(tea.StringValue(result.Body.Cert)),
		Daemon:            tea.StringValue(result.Body.Common),
		ExpireTime:        tea.StringValue(result.Body.EndDate),
		Expired:           tea.BoolValue(result.Body.Expired),
		IssuerCertificate: []byte(tea.StringValue(result.Body.Issuer)),
		KeyPem:            []byte(tea.StringValue(result.Body.Key)),
	}, err
}
func (p *AliProvider) SyncCertificate() error {
	return nil
}

func (p *AliProvider) UploadCertificate(req UploadCertificateRequest) (int64, error) {
	client, err := CreateClient(p)
	if err != nil {
		return 0, err
	}
	request := &cas.UploadUserCertificateRequest{}
	request.Name = req.Name
	request.Cert = req.Crt
	request.Key = req.Key
	response, err := client.UploadUserCertificate(request)
	if err != nil {
		return 0, err
	}
	return *response.Body.CertId, nil
}
func (p *AliProvider) ApplyCertificate(instanceId string) error {
	return nil
}

func (p *AliProvider) DeleteCertificate(certId int64) error {
	client, err := CreateClient(p)
	if err != nil {
		return err
	}

	response, err := client.DeleteUserCertificate(&cas.DeleteUserCertificateRequest{CertId: tea.Int64(certId)})
	if err != nil {
		return err
	}
	if *response.StatusCode != 200 {
		return fmt.Errorf("删除证书失败，错误码为：%v", *response.StatusCode)
	}
	return nil
}

func CreateClient(p *AliProvider) (*cas.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     tea.String(p.AccessKey),
		AccessKeySecret: tea.String(p.SecretKey),
		Endpoint:        tea.String("cas.aliyuncs.com"),
	}
	client, err := cas.NewClient(config)
	return client, err
}
