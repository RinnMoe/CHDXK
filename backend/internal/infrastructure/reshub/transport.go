package reshub

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"jcourse/config"
	"net/http"
	"os"
	"strings"
)

func newTransport(conf config.ResHubConfig) (*http.Transport, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	certFile, keyFile, caFile := strings.TrimSpace(conf.ClientCertFile), strings.TrimSpace(conf.ClientKeyFile), strings.TrimSpace(conf.CAFile)
	if (certFile == "") != (keyFile == "") {
		return nil, fmt.Errorf("res-hub client certificate and key must be configured together")
	}
	if certFile == "" && caFile == "" {
		return transport, nil
	}
	if !strings.HasPrefix(strings.TrimSpace(conf.Issuer), "https://") {
		return nil, fmt.Errorf("res-hub client TLS configuration requires an HTTPS issuer")
	}
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	if certFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, fmt.Errorf("load res-hub client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}
	if caFile != "" {
		data, err := os.ReadFile(caFile)
		if err != nil {
			return nil, fmt.Errorf("read res-hub CA file: %w", err)
		}
		roots, err := x509.SystemCertPool()
		if err != nil {
			roots = x509.NewCertPool()
		}
		if !roots.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("res-hub CA file contains no valid certificates")
		}
		tlsConfig.RootCAs = roots
	}
	transport.TLSClientConfig = tlsConfig
	return transport, nil
}
