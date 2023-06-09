package aliacr

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

func getMockAdapter(t *testing.T, hasCred, health bool) (*adapter, *httptest.Server) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				if health {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusBadRequest)
				}
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				w.WriteHeader(http.StatusOK)
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodPost,
			Pattern: "/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				if buf, e := io.ReadAll(&io.LimitedReader{R: r.Body, N: 80}); e == nil {
					fmt.Println("\t", string(buf))
				}
				w.WriteHeader(http.StatusOK)
			},
		},
	)

	registry := &model.Registry{
		Type: model.RegistryTypeAliAcr,
		URL:  server.URL,
	}
	if hasCred {
		registry.Credential = &model.Credential{
			AccessKey:    "MockAccessKey",
			AccessSecret: "MockAccessSecret",
		}
	}
	return &adapter{
		Adapter:  native.NewAdapter(registry),
		region:   "test-region",
		domain:   server.URL,
		registry: registry,
	}, server
}

func TestAdapter_Info(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	info, err := a.Info()
	assert.Nil(t, err)
	assert.NotNil(t, info)

	assert.EqualValues(t, 1, len(info.SupportedResourceTypes))
	assert.EqualValues(t, model.ResourceTypeImage, info.SupportedResourceTypes[0])
}

func Test_getRegion(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		wantRegion string
		wantErr    bool
	}{
		{"registry shanghai", "https://registry.cn-shanghai.aliyuncs.com", "cn-shanghai", false},
		{"invalid registry shanghai", "http://registry.cn-shanghai.aliyuncs.com", "", true},
		{"registry hangzhou", "https://registry.cn-hangzhou.aliyuncs.com", "cn-hangzhou", false},
		{"registry hangzhou vpc", "https://registry-vpc.cn-hangzhou.aliyuncs.com", "cn-hangzhou", false},
		{"registry hangzhou internal", "https://registry-internal.cn-hangzhou.aliyuncs.com", "cn-hangzhou", false},
		{"cr shanghai", "https://cr.cn-shanghai.aliyuncs.com", "cn-shanghai", false},
		{"cr hangzhou", "https://cr.cn-hangzhou.aliyuncs.com", "cn-hangzhou", false},
		{"invalid cr url", "https://acr.cn-hangzhou.aliyuncs.com", "", true},
		{"invalid registry url", "https://registry.cn-hangzhou.ali.com", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRegion, err := getRegion(tt.url)
			if tt.wantErr {
				assert.NotNil(t, err)
			}
			assert.Equal(t, tt.wantRegion, gotRegion)
		})
	}
}

var urlForBenchmark = []string{
	"https://cr.cn-hangzhou.aliyuncs.com",
	"https://registry.cn-shanghai.aliyuncs.com",
	"https://registry-vpc.cn-shanghai.aliyuncs.com",
	"https://registry-internal.cn-shanghai.aliyuncs.com",
}

func BenchmarkGetRegion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, url := range urlForBenchmark {
			getRegion(url)
		}
	}
}

func Test_adapter_FetchArtifacts(t *testing.T) {
	a, s := getMockAdapter(t, true, true)
	defer s.Close()
	var filters = []*model.Filter{}
	var resources, err = a.FetchArtifacts(filters)
	assert.NotNil(t, err)
	assert.Nil(t, resources)
}
func Test_aliyunAuthCredential_isCacheTokenValid(t *testing.T) {
	type fields struct {
		region              string
		accessKey           string
		secretKey           string
		cacheToken          *registryTemporaryToken
		cacheTokenExpiredAt time.Time
	}

	var nilTime time.Time
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{"nil cacheTokenExpiredAt", fields{"test-region", "MockAccessKey", "MockSecretKey", nil, nilTime}, false},
		{"nil cacheToken", fields{"test-region", "MockAccessKey", "MockSecretKey", nil, time.Time{}}, false},
		{"expired", fields{"test-region", "MockAccessKey", "MockSecretKey", &registryTemporaryToken{}, time.Now().AddDate(0, 0, -1)}, false},
		{"ok", fields{"test-region", "MockAccessKey", "MockSecretKey", &registryTemporaryToken{}, time.Now().AddDate(0, 0, 1)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &aliyunAuthCredential{
				region:              tt.fields.region,
				accessKey:           tt.fields.accessKey,
				secretKey:           tt.fields.secretKey,
				cacheToken:          tt.fields.cacheToken,
				cacheTokenExpiredAt: tt.fields.cacheTokenExpiredAt,
			}
			if got := a.isCacheTokenValid(); got != tt.want {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}
