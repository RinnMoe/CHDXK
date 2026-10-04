package reshub

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"jcourse/config"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestClientSendsCertificateAndVerifiesPrivateServerCA(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "chdxk-test-client"}, NotBefore: time.Now().Add(-time.Minute), NotAfter: time.Now().Add(time.Hour), IsCA: true, BasicConstraintsValid: true, KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certPath, keyPath, caPath := filepath.Join(dir, "client.pem"), filepath.Join(dir, "client.key"), filepath.Join(dir, "server-ca.pem")
	if err := os.WriteFile(certPath, certPEM, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatal(err)
	}
	clientRoots := x509.NewCertPool()
	clientRoots.AppendCertsFromPEM(certPEM)
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.TLS.PeerCertificates) != 1 || r.TLS.PeerCertificates[0].Subject.CommonName != "chdxk-test-client" {
			t.Error("missing verified client certificate")
			w.WriteHeader(403)
			return
		}
		w.WriteHeader(200)
	}))
	server.TLS = &tls.Config{ClientAuth: tls.RequireAndVerifyClientCert, ClientCAs: clientRoots, MinVersion: tls.VersionTLS12}
	server.StartTLS()
	defer server.Close()
	if err := os.WriteFile(caPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw}), 0600); err != nil {
		t.Fatal(err)
	}
	client := mustNewClient(t, config.ResHubConfig{Issuer: server.URL, ClientCertFile: certPath, ClientKeyFile: keyPath, CAFile: caPath})
	response, err := client.httpClient.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("status %d", response.StatusCode)
	}
	withoutCert := mustNewClient(t, config.ResHubConfig{Issuer: server.URL, CAFile: caPath})
	if response, err := withoutCert.httpClient.Get(server.URL); err == nil {
		response.Body.Close()
		t.Fatal("mTLS accepted a client without certificate")
	}
}

func TestClientRejectsInvalidTLSConfiguration(t *testing.T) {
	for _, conf := range []config.ResHubConfig{
		{Issuer: "https://hub.test", ClientCertFile: "missing.pem"},
		{Issuer: "http://hub.test", ClientCertFile: "missing.pem", ClientKeyFile: "missing.key"},
		{Issuer: "https://hub.test", ClientCertFile: "missing.pem", ClientKeyFile: "missing.key"},
		{Issuer: "https://hub.test", CAFile: "missing.pem"},
	} {
		if _, err := NewClient(conf); err == nil {
			t.Fatalf("invalid TLS config accepted: %+v", conf)
		}
	}
}
