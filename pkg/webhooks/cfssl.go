package webhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/cloudflare/cfssl/api/client"
	"github.com/cloudflare/cfssl/cli"
	"github.com/cloudflare/cfssl/cli/serve"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/helpers"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/log"
	ocspconfig "github.com/cloudflare/cfssl/ocsp/config"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// CFSSLServer exposes methods for launching an embedded CFSSL server,
// retrieving its CA info, and requesting signed certs from it.
type CFSSLServer struct {
	ca    []byte
	caKey []byte
}

// NewCFSSLServer constructs a CFSSLServer instance with the given configuration.
// It takes an optional PEM-encoded CA and CA key used by the server.
// If a CA and CA key are not provided, they will be generated and used to launch the server.
func NewCFSSLServer(ca, caKey []byte) (*CFSSLServer, error) {

	// Wrap CFSSL's logger in our custom implementation
	log.SetLogger(&cfsslLogger{ctrl.Log.WithName("cfssl")})
	log.Level = log.LevelInfo

	var err error

	if len(ca) == 0 || len(caKey) == 0 {
		logger.Info("CA and CA key not provided; initializing CA")
		ca, _, caKey, err = initca.New(&csr.CertificateRequest{
			CN:         "greymatter.io",
			Hosts:      []string{"greymatter.io"},
			KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
			CA:         &csr.CAConfig{Expiry: "8760h"},
		})
		if err != nil {
			return nil, err
		}
	} else {
		logger.Info("Using provided CA and CA key")
	}

	if _, err := helpers.ParseCertificatesPEM(ca); err != nil {
		err = fmt.Errorf("failed to decode PEM block")
		logger.Error(err, "Detected invalid CA")
		return nil, err
	}
	if _, err := helpers.ParsePrivateKeyPEM(caKey); err != nil {
		err = fmt.Errorf("failed to decode PEM block")
		logger.Error(err, "Detected invalid CA key")
		return nil, err
	}

	return &CFSSLServer{
		ca:    ca,
		caKey: caKey,
	}, nil
}

func (cs *CFSSLServer) Start() error {
	os.Setenv("CFSSL_CA", string(cs.ca))
	os.Setenv("CFSSL_CA_KEY", string(cs.caKey))

	logger.Info("Launching CA server for issuing signed certs")

	go func() {
		err := serve.Command.Main(nil, cli.Config{
			Port: 8888,
			// Disable endpoints except the ones we use.
			// ref: https://github.com/cloudflare/cfssl/blob/master/cli/serve/serve.go#L121
			Disable:   "sign,authsign,crl,gencrl,bundle,newkey,init_ca,scan,scaninfo,certinfo,ocspsign,revoke,/,health,certadd",
			CAFile:    "env:CFSSL_CA",
			CAKeyFile: "env:CFSSL_CA_KEY",
			CFG: &config.Config{
				Signing: &config.Signing{
					Default: &config.SigningProfile{
						Usage: []string{
							"signing",
							"key encipherment",
							"server auth",
							"client auth",
						},
						Expiry: time.Hour * 8760,
					},
				},
				OCSP: &ocspconfig.Config{},
			},
		})
		if err != nil {
			logger.Error(err, "Failed to serve CFSSL server")
		}
	}()

	// Ensure our CFSSL server is running and able to issue certs.
	// We assume it should take no longer than 5 seconds to initialize.
	remote := client.NewServer("http://127.0.0.1:8888")
	remote.SetRequestTimeout(time.Second)
	timer := time.NewTimer(time.Second * 5)
	for {
		select {
		case <-timer.C:
			logger.Error(context.DeadlineExceeded, "CFSSL server failed to initialize")
			return context.DeadlineExceeded
		default:
			if _, err := remote.Info([]byte(`{}`)); err == nil {
				return nil
			}
		}
	}
}

func (cs *CFSSLServer) GetCABundle() []byte {
	return cs.ca
}

func (cs *CFSSLServer) RequestCert(req csr.CertificateRequest) ([]byte, []byte, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, nil, err
	}

	c := http.Client{Timeout: time.Second * 3}
	resp, err := getCFSSLResponse(c, "newcert", fmt.Sprintf(`{"request":%s}`, string(reqBytes)))
	if err != nil {
		return nil, nil, err
	}

	if !resp.Success {
		return nil, nil, fmt.Errorf("server returned failure response")
	}

	if len(resp.Result.Cert) == 0 {
		err := fmt.Errorf("server response did not contain certificate")
		logger.Error(err, "failed to retrieve cert")
		return nil, nil, err
	}

	if len(resp.Result.Key) == 0 {
		err := fmt.Errorf("server response did not contain private_key")
		logger.Error(err, "failed to retrieve cert key")
		return nil, nil, err
	}

	return []byte(resp.Result.Cert), []byte(resp.Result.Key), nil
}

// models a subset of fields returned in the response from the CFSSL server
type cfsslResponse struct {
	Success bool `json:"success"`
	Result  struct {
		Cert string `json:"certificate"`
		Key  string `json:"private_key"`
	} `json:"result"`
}

func getCFSSLResponse(c http.Client, path, data string) (*cfsslResponse, error) {
	url := fmt.Sprintf("http://127.0.0.1:8888/api/v1/cfssl/%s", path)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return nil, err
	}
	httpResp, err := c.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	httpRespBody, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}
	resp := &cfsslResponse{}
	if err := json.Unmarshal(httpRespBody, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// implements the cfssl/log.SysLogWriter interface
type cfsslLogger struct {
	logger logr.Logger
}

func (cl *cfsslLogger) Debug(msg string) {
	cl.logger.Info(msg, "level", "debug")
}

func (cl *cfsslLogger) Info(msg string) {
	// Suppress noisy info logs we don't care about
	if strings.Contains(msg, "is explicitly disabled") {
		return
	}
	cl.logger.Info(msg)
}

func (cl *cfsslLogger) Warning(msg string) {
	// Suppress the ocsp signer warning since we aren't using it
	if strings.Contains(msg, "couldn't initialize ocsp signer") {
		return
	}
	cl.logger.Info(msg, "level", "warn")
}

func (cl *cfsslLogger) Err(msg string) {
	cl.logger.Error(fmt.Errorf("%s", msg), "", "level", "error")
}

func (cl *cfsslLogger) Crit(msg string) {
	cl.logger.Error(fmt.Errorf("%s", msg), "", "level", "critical")
}

func (cl *cfsslLogger) Emerg(msg string) {
	cl.logger.Error(fmt.Errorf("%s", msg), "", "level", "fatal")
}
