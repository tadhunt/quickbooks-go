package quickbooks

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeDump(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "bearer header redacted, scheme kept",
			in:   "GET /v3/company/123/query HTTP/1.1\r\nAuthorization: Bearer eyJhbGciOi.secret.value\r\nAccept: application/json\r\n\r\n",
			want: "GET /v3/company/123/query HTTP/1.1\r\nAuthorization: Bearer [REDACTED]\r\nAccept: application/json\r\n\r\n",
		},
		{
			name: "basic header redacted",
			in:   "POST /oauth2/v1/tokens/bearer HTTP/1.1\r\nAuthorization: Basic Y2xpZW50OnNlY3JldA==\r\n\r\n",
			want: "POST /oauth2/v1/tokens/bearer HTTP/1.1\r\nAuthorization: Basic [REDACTED]\r\n\r\n",
		},
		{
			name: "lowercase header name still matched",
			in:   "GET / HTTP/1.1\r\nauthorization: bearer abc123\r\n\r\n",
			want: "GET / HTTP/1.1\r\nAuthorization: bearer [REDACTED]\r\n\r\n",
		},
		{
			name: "token response body redacted",
			in:   "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"access_token\":\"AB11\",\"refresh_token\":\"RT22\",\"id_token\":\"ID33\",\"expires_in\":3600}",
			want: "HTTP/1.1 200 OK\r\nContent-Type: application/json\r\n\r\n{\"access_token\":\"[REDACTED]\",\"refresh_token\":\"[REDACTED]\",\"id_token\":\"[REDACTED]\",\"expires_in\":3600}",
		},
		{
			name: "token fields with whitespace around colon",
			in:   `{"access_token" : "AB11"}`,
			want: `{"access_token" : "[REDACTED]"}`,
		},
		{
			name: "query response body left intact",
			in:   "HTTP/1.1 200 OK\r\n\r\n{\"QueryResponse\":{\"Invoice\":[{\"Id\":\"59\"}],\"maxResults\":50,\"totalCount\":50}}",
			want: "HTTP/1.1 200 OK\r\n\r\n{\"QueryResponse\":{\"Invoice\":[{\"Id\":\"59\"}],\"maxResults\":50,\"totalCount\":50}}",
		},
		{
			name: "header without a credential does not panic",
			in:   "GET / HTTP/1.1\r\nAuthorization:\r\n\r\n",
			want: "GET / HTTP/1.1\r\nAuthorization:  [REDACTED]\r\n\r\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(sanitizeDump([]byte(tt.in)))
			assert.Equal(t, tt.want, got)

			// Whatever else happens, no dump may retain a credential.
			for _, secret := range []string{"eyJhbGciOi.secret.value", "Y2xpZW50OnNlY3JldA==", "abc123", "AB11", "RT22", "ID33"} {
				assert.NotContains(t, got, secret)
			}
		})
	}
}

func TestSanitizeDumpPreservesLineCount(t *testing.T) {
	in := "GET / HTTP/1.1\r\nAuthorization: Bearer abc\r\nAccept: application/json\r\n\r\nbody\n"
	got := string(sanitizeDump([]byte(in)))

	assert.Equal(t, strings.Count(in, "\n"), strings.Count(got, "\n"))
}

func TestPageSize(t *testing.T) {
	c := &Client{}
	assert.Equal(t, queryPageSize, c.pageSize(), "zero value must fall back to the default")

	c.SetQueryPageSize(20)
	assert.Equal(t, 20, c.pageSize())

	c.SetQueryPageSize(0)
	assert.Equal(t, queryPageSize, c.pageSize(), "0 restores the default")

	c.SetQueryPageSize(-5)
	assert.Equal(t, queryPageSize, c.pageSize(), "negative is treated as unset")
}
