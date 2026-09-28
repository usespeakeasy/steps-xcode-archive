package step

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/bitrise-io/go-utils/v2/log"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnect"
	"github.com/bitrise-io/go-xcode/v2/autocodesign/devportalclient/appstoreconnectclient"
	"github.com/stretchr/testify/require"
)

func TestCertificateQueriesMatchAccountType(t *testing.T) {
	for _, tt := range []struct {
		name       string
		enterprise bool
		host       string
		types      []string
	}{
		{
			name:       "enterprise excludes unsupported certificate types",
			enterprise: true,
			host:       "api.enterprise.developer.apple.com",
			types:      []string{"IOS_DEVELOPMENT", "IOS_DISTRIBUTION"},
		},
		{
			name:  "app store keeps all certificate types",
			host:  "api.appstoreconnect.apple.com",
			types: []string{"DEVELOPMENT", "IOS_DEVELOPMENT", "DISTRIBUTION", "IOS_DISTRIBUTION"},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			httpClient := &certificateQueryClient{t: t, host: tt.host}
			client := appstoreconnect.NewClient(httpClient, "", "", nil, tt.enterprise,
				log.NewLogger(), appstoreconnect.NoOpAnalyticsTracker{})
			certificates, err := appstoreconnectclient.NewCertificateSource(client).QueryAllIOSCertificates()
			require.NoError(t, err)
			require.Equal(t, tt.types, httpClient.types)
			require.Len(t, certificates, len(tt.types))
			for _, certificateType := range tt.types {
				require.Contains(t, certificates, appstoreconnect.CertificateType(certificateType))
			}
		})
	}
}

func TestEnterpriseCertificateQueryPreservesAPIErrors(t *testing.T) {
	httpClient := &certificateQueryClient{
		t: t, host: "api.enterprise.developer.apple.com", reject: true,
	}
	client := appstoreconnect.NewClient(httpClient, "", "", nil, true,
		log.NewLogger(), appstoreconnect.NoOpAnalyticsTracker{})
	certificates, err := appstoreconnectclient.NewCertificateSource(client).QueryAllIOSCertificates()
	require.Error(t, err)
	require.Empty(t, certificates)
	require.Equal(t, []string{"IOS_DEVELOPMENT"}, httpClient.types)
}

type certificateQueryClient struct {
	t      *testing.T
	host   string
	types  []string
	reject bool
}

func (c *certificateQueryClient) Do(req *http.Request) (*http.Response, error) {
	require.Equal(c.t, http.MethodGet, req.Method)
	require.Equal(c.t, c.host, req.URL.Host)
	require.Equal(c.t, "/v1/certificates", req.URL.Path)
	c.types = append(c.types, req.URL.Query().Get("filter[certificateType]"))
	status := http.StatusOK
	body := `{"data":[],"links":{},"meta":{"paging":{"total":0}}}`
	if c.reject {
		status = http.StatusForbidden
		body = `{"errors":[{"status":"403","code":"FORBIDDEN","title":"Access denied"}]}`
	}
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}
