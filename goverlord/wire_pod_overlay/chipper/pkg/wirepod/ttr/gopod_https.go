// goverlord/wire_pod_overlay/chipper/pkg/wirepod/ttr/gopod_https.go
package wirepod_ttr

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kercre123/wire-pod/chipper/pkg/logger"
	"github.com/kercre123/wire-pod/chipper/pkg/vars"
)

// GOPOD as a phone app (2026-10-24, FINISH_PLAN_001.md C3; operator: "Go on C3 for
// my phone"): a phone offers "Install app" only for a secure (https) page, so
// Wire-Pod also serves GOPOD securely on app.https_port - the same pages and
// the same GOPOD, nothing copied. The certificate is GOPOD's own, made here: a
// local GOPOD authority (kept for years - the phone trusts it once, from
// /gopod_ca.crt) and a server certificate for this machine's names and network
// addresses, made again whenever they change (on the road, a new network). All
// of it lives in app.cert_dir (private, the operator's home, never the repo).

func gopodCertDir() string { return gopodHomePath(gopodBrobotsSettings.App.CertDir) }

// gopodMachineNames: this machine's names and network addresses, sorted.
func gopodMachineNames() []string {
	names := map[string]bool{"localhost": true, "127.0.0.1": true}
	if h, err := os.Hostname(); err == nil && h != "" {
		names[h] = true
		names[h+".local"] = true
	}
	if addrs, err := net.InterfaceAddrs(); err == nil {
		for _, a := range addrs {
			if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil && !n.IP.IsLoopback() {
				names[n.IP.String()] = true
			}
		}
	}
	var out []string
	for n := range names {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func gopodWritePEM(path, kind string, der []byte, private bool) error {
	mode := os.FileMode(0o644)
	if private {
		mode = 0o600
	}
	if err := gopodWriteFile(path, pem.EncodeToMemory(&pem.Block{Type: kind, Bytes: der}), mode); err != nil {
		return err
	}
	if _, uid, gid := gopodRecordRoot(); uid > 0 {
		os.Chown(path, uid, gid)
	}
	return nil
}

func gopodReadPEM(path string) []byte {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	if b, _ := pem.Decode(data); b != nil {
		return b.Bytes
	}
	return nil
}

// gopodEnsureCerts: GOPOD's authority (made once) and this machine's server
// certificate (made again when its names change). The server cert and key paths.
func gopodEnsureCerts() (string, string, error) {
	dir := gopodCertDir()
	if dir == "" {
		return "", "", errors.New("app.cert_dir is empty")
	}
	if err := gopodMkdirAll(dir, 0o700); err != nil {
		return "", "", err
	}
	if _, uid, gid := gopodRecordRoot(); uid > 0 {
		os.Chown(dir, uid, gid)
	}
	caCrt, caKey := filepath.Join(dir, "ca.crt"), filepath.Join(dir, "ca.key")
	srvCrt, srvKey, namesFile := filepath.Join(dir, "server.crt"), filepath.Join(dir, "server.key"), filepath.Join(dir, "server.names")
	name := strings.TrimSpace(gopodBrobotsSettings.App.CAName)
	if name == "" {
		name = "GOPOD local"
	}
	var ca *x509.Certificate
	var caPriv *ecdsa.PrivateKey
	if der := gopodReadPEM(caCrt); der != nil {
		if c, err := x509.ParseCertificate(der); err == nil {
			if k, err := x509.ParseECPrivateKey(gopodReadPEM(caKey)); err == nil {
				ca, caPriv = c, k
			}
		}
	}
	// 2026-10-31 (operator: "Do/finish up * CA fencing"): an authority without its
	// fence is replaced by a fenced one - the old one kept aside, never deleted;
	// phones trust the new one once.
	dnsFence, ipFence := gopodCAFence()
	if ca != nil && (len(dnsFence) > 0 || len(ipFence) > 0) && len(ca.PermittedDNSDomains) == 0 && len(ca.PermittedIPRanges) == 0 {
		stamp := time.Now().Format("20060102_150405")
		os.Rename(caCrt, filepath.Join(dir, "ca_unfenced_"+stamp+".crt"))
		os.Rename(caKey, filepath.Join(dir, "ca_unfenced_"+stamp+".key"))
		logger.Println("GOPOD_APP the certificate authority had no fence - kept aside as ca_unfenced_" + stamp + "; a fenced one is made (phones trust /gopod_ca.crt again)")
		ca, caPriv = nil, nil
	}
	if ca == nil {
		k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return "", "", err
		}
		serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
		tmpl := &x509.Certificate{
			SerialNumber: serial, Subject: pkix.Name{CommonName: name, Organization: []string{"GOPOD"}},
			NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(10, 0, 0),
			IsCA: true, BasicConstraintsValid: true, MaxPathLenZero: true,
			KeyUsage: x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
			// The fence: it may only ever vouch for these (app.fence_dns, app.fence_ips).
			PermittedDNSDomainsCritical: true, PermittedDNSDomains: dnsFence, PermittedIPRanges: ipFence,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &k.PublicKey, k)
		if err != nil {
			return "", "", err
		}
		kder, _ := x509.MarshalECPrivateKey(k)
		if err := gopodWritePEM(caKey, "EC PRIVATE KEY", kder, true); err != nil {
			return "", "", err
		}
		if err := gopodWritePEM(caCrt, "CERTIFICATE", der, false); err != nil {
			return "", "", err
		}
		ca, _ = x509.ParseCertificate(der)
		caPriv = k
		os.Remove(namesFile)
		logger.Println("GOPOD_APP made GOPOD's own certificate authority (" + name + ")")
	}
	// Only the names inside the fence go in the server certificate (a browser
	// refuses a certificate naming anything outside it).
	var names []string
	for _, n := range gopodMachineNames() {
		if gopodInsideFence(n, ca) {
			names = append(names, n)
		}
	}
	want := strings.Join(names, "\n")
	if old, err := os.ReadFile(namesFile); err == nil && string(old) == want {
		if der := gopodReadPEM(srvCrt); der != nil {
			if c, err := x509.ParseCertificate(der); err == nil && time.Until(c.NotAfter) > 30*24*time.Hour {
				return srvCrt, srvKey, nil
			}
		}
	}
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 62))
	tmpl := &x509.Certificate{
		SerialNumber: serial, Subject: pkix.Name{CommonName: "GOPOD", Organization: []string{"GOPOD"}},
		NotBefore: time.Now().Add(-time.Hour), NotAfter: time.Now().AddDate(0, 0, 397),
		KeyUsage: x509.KeyUsageDigitalSignature, ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	for _, n := range names {
		if ip := net.ParseIP(n); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, n)
		}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, ca, &k.PublicKey, caPriv)
	if err != nil {
		return "", "", err
	}
	kder, _ := x509.MarshalECPrivateKey(k)
	if err := gopodWritePEM(srvKey, "EC PRIVATE KEY", kder, true); err != nil {
		return "", "", err
	}
	if err := gopodWritePEM(srvCrt, "CERTIFICATE", der, false); err != nil {
		return "", "", err
	}
	gopodWriteFile(namesFile, []byte(want), 0o600)
	logger.Println("GOPOD_APP made the server certificate for " + strconv.Itoa(len(names)) + " names and addresses of this machine")
	return srvCrt, srvKey, nil
}

// GopodStartHTTPS: GOPOD, securely, on app.https_port (the same pages, the same
// handlers). Started beside Wire-Pod's own web server; a failure is logged and
// the plain page goes on as ever.
func GopodStartHTTPS() {
	port := gopodBrobotsSettings.App.HTTPSPort
	if port <= 0 {
		return
	}
	crt, key, err := gopodEnsureCerts()
	if err != nil {
		logger.Println("GOPOD_APP no secure page: " + err.Error())
		return
	}
	logger.Println("GOPOD_APP secure page on port " + strconv.Itoa(port) + " (the phone trusts /gopod_ca.crt once)")
	if err := http.ListenAndServeTLS(":"+strconv.Itoa(port), crt, key, vars.GopodGated(http.DefaultServeMux)); err != nil {
		logger.Println("GOPOD_APP secure page stopped: " + err.Error())
	}
}

func init() {
	// GOPOD's authority, for a phone to trust once (public - no key in it).
	http.HandleFunc("/gopod_ca.crt", func(w http.ResponseWriter, r *http.Request) {
		data, err := os.ReadFile(filepath.Join(gopodCertDir(), "ca.crt"))
		if err != nil {
			http.Error(w, "not made yet", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/x-x509-ca-cert")
		w.Header().Set("Content-Disposition", `attachment; filename="gopod_ca.crt"`)
		w.Write(data)
	})
}

// gopodCAFence: the fence from the settings - app.fence_dns (and this machine's
// own hostname), app.fence_ips as networks.
func gopodCAFence() (dns []string, ips []*net.IPNet) {
	a := gopodBrobotsSettings.App
	seen := map[string]bool{}
	add := func(d string) {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" && !seen[d] {
			seen[d] = true
			dns = append(dns, d)
		}
	}
	for _, d := range a.FenceDNS {
		add(d)
	}
	if len(dns) > 0 {
		if h, err := os.Hostname(); err == nil {
			add(h)
		}
	}
	for _, c := range a.FenceIPs {
		if _, n, err := net.ParseCIDR(strings.TrimSpace(c)); err == nil {
			ips = append(ips, n)
		}
	}
	return dns, ips
}

// gopodInsideFence: a name or address the authority may vouch for (no fence:
// any).
func gopodInsideFence(name string, ca *x509.Certificate) bool {
	if ca == nil || (len(ca.PermittedDNSDomains) == 0 && len(ca.PermittedIPRanges) == 0) {
		return true
	}
	if ip := net.ParseIP(name); ip != nil {
		for _, n := range ca.PermittedIPRanges {
			if n.Contains(ip) {
				return true
			}
		}
		return false
	}
	name = strings.ToLower(name)
	for _, d := range ca.PermittedDNSDomains {
		d = strings.ToLower(strings.TrimPrefix(d, "."))
		if name == d || strings.HasSuffix(name, "."+d) {
			return true
		}
	}
	return false
}

// ## EOF
