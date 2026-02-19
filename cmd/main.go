package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	apisecurityv1 "gofronet-foundation/gofro-control/gen/go/api/security/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	rootCAPath   = "./data/certs/root-ca.crt"
	nodeKeyPath  = "./data/certs/node.key"
	nodeCertPath = "./data/certs/node.crt"
	nodeCertDER  = "./data/certs/node.crt.der"
)

func main() {
	var (
		bootstrapToken = flag.String("bootstrap", "", "Bootstrap token")
		controlAddr    = flag.String("control-address", "", "Control plane address host:port")
		timeout        = flag.Duration("timeout", 20*time.Second, "Bootstrap timeout")
	)
	flag.Parse()

	fmt.Println(*controlAddr)

	if *bootstrapToken == "" || *controlAddr == "" {
		fmt.Fprintln(os.Stderr, "usage: node --bootstrap <token> --control-address <host:port>")
		os.Exit(2)
	}

	if err := runBootstrap(*bootstrapToken, *controlAddr, *timeout); err != nil {
		fmt.Fprintln(os.Stderr, "bootstrap failed:", err)
		os.Exit(1)
	}

	fmt.Println("bootstrap OK")
}

func runBootstrap(token, controlAddr string, timeout time.Duration) error {
	// 0) Root CA
	rootPEM, err := os.ReadFile(rootCAPath)
	if err != nil {
		return fmt.Errorf("read root ca %s: %w", rootCAPath, err)
	}
	roots := x509.NewCertPool()
	if ok := roots.AppendCertsFromPEM(rootPEM); !ok {
		return fmt.Errorf("failed to parse root ca pem: %s", rootCAPath)
	}

	// 1) TLS config for bootstrap (server-auth only)
	host, _, err := net.SplitHostPort(controlAddr)
	if err != nil {
		// allow passing without port? but you said host:port, so be strict:
		return fmt.Errorf("control-address must be host:port: %w", err)
	}
	tlsCfg := &tls.Config{
		RootCAs:    roots,
		MinVersion: tls.VersionTLS12,
		// IMPORTANT: this is used for name verification (DNS or IP SAN).
		ServerName: host,
	}

	// 2) dial gRPC
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := grpc.NewClient(
		controlAddr,
		grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)),
	)
	if err != nil {
		return fmt.Errorf("dial control plane: %w", err)
	}
	defer conn.Close()

	client := apisecurityv1.NewBootstrapServiceClient(conn)

	// 3) generate key + CSR DER
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	csrDER, err := makeCSRDER(priv)
	if err != nil {
		return fmt.Errorf("create csr: %w", err)
	}

	// 4) call Bootstrap
	resp, err := client.Bootstrap(ctx, &apisecurityv1.BootstrapRequest{
		BootstrapToken: token,
		CsrDer:         csrDER,
	})
	if err != nil {
		return fmt.Errorf("bootstrap rpc: %w", err)
	}

	if len(resp.LeafCertDer) == 0 {
		return errors.New("empty leaf_cert_der in response")
	}

	// 5) persist key + certs
	if err := os.MkdirAll(filepath.Dir(nodeKeyPath), 0o700); err != nil {
		return fmt.Errorf("mkdir cert dir: %w", err)
	}

	// key: PKCS#8 PEM
	keyPEM, err := marshalPrivateKeyPKCS8PEM(priv)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}
	if err := os.WriteFile(nodeKeyPath, keyPEM, 0o600); err != nil {
		return fmt.Errorf("write node key: %w", err)
	}

	// cert: convert DER -> PEM for easy tls.X509KeyPair later
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: resp.LeafCertDer})
	if err := os.WriteFile(nodeCertPath, certPEM, 0o644); err != nil {
		return fmt.Errorf("write node cert pem: %w", err)
	}

	// optional: keep raw der too
	_ = os.WriteFile(nodeCertDER, resp.LeafCertDer, 0o644)

	fmt.Printf("node_id=%s expires_unix=%d\n", resp.NodeId, resp.ExpiresUnix)
	fmt.Printf("wrote %s and %s\n", nodeKeyPath, nodeCertPath)

	return nil
}

func makeCSRDER(priv *ecdsa.PrivateKey) ([]byte, error) {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "node"
	}

	// SAN must be non-empty, у тебя это проверяется на control plane.
	// Базовый вариант: DNS SAN = hostname, + IP SAN из интерфейсов (если найдём).
	dnsSAN := []string{hostname}
	ipSAN := localIPs()

	tmpl := &x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: hostname, // можно, но на сервере всё равно лучше формировать node_id самому
		},
		DNSNames:    dnsSAN,
		IPAddresses: ipSAN,
	}

	return x509.CreateCertificateRequest(rand.Reader, tmpl, priv)
}

func localIPs() []net.IP {
	var out []net.IP

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if (iface.Flags & net.FlagUp) == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range addrs {
			var ip net.IP
			switch v := a.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // пропускаем IPv6 (можешь расширить)
			}
			if ip.IsLoopback() {
				continue
			}
			out = append(out, ip)
		}
	}
	return out
}

func marshalPrivateKeyPKCS8PEM(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}
