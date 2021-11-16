package webhooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/cloudflare/cfssl/api/client"
	"github.com/cloudflare/cfssl/cli"
	"github.com/cloudflare/cfssl/cli/serve"
	"github.com/cloudflare/cfssl/config"
	"github.com/cloudflare/cfssl/csr"
	"github.com/cloudflare/cfssl/initca"
	"github.com/cloudflare/cfssl/log"
	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

// serveCFSSL launches a CFSSL server for issuing TLS certs:
// 1. for the operator's webhook server, to admit requests from the Kubelet
// 2. TODO: for the operator to issue intermediate CAs for intra-mesh communication (if desired)
// For now, serveCFSSL is local to the webhooks package. This will likely move into its own package later.
// The returned bytes contain the CA returned from a running server, ensuring the server is initialized.
func serveCFSSL() ([]byte, error) {

	// Wrap CFSSL's logs in our custom logger
	log.SetLogger(&cfsslLogger{ctrl.Log.WithName("cfssl")})
	log.Level = log.LevelInfo

	logger.Info("Initializing CA")

	// Initialize our CA for signing and issuing certs
	// https://github.com/cloudflare/cfssl/blob/master/cli/gencert/gencert.go#L66-L93
	caCert, _, caKey, err := initca.New(&csr.CertificateRequest{
		CN:         "greymatter.io",
		Hosts:      []string{"greymatter.io"},
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
		CA:         &csr.CAConfig{Expiry: "8760h"},
	})
	if err != nil {
		return nil, err
	}

	os.Setenv("CFSSL_CA", string(caCert))
	os.Setenv("CFSSL_CA_KEY", string(caKey))

	logger.Info("Launching CA server for issuing signed certs")

	go func() {
		err := serve.Command.Main(nil, cli.Config{
			Port: 8888,
			// Disable endpoints expect the ones we use
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
			},
		})
		if err != nil {
			logger.Error(err, "Failed to serve CFSSL server")
		}
	}()

	// Ensure our CFSSL server is running and able to issue certs.
	remote := client.NewServer("http://127.0.0.1:8888")
	remote.SetRequestTimeout(time.Second * 2)
	var caBundle string
	for {
		info, err := remote.Info([]byte(`{}`))
		if err != nil {
			time.Sleep(time.Second * 3)
			continue
		}
		caBundle = info.Certificate
		break
	}

	return []byte(caBundle), nil
}

func requestWebhookCerts() ([]string, error) {
	req := &csr.CertificateRequest{
		CN:         "admission",
		Hosts:      []string{"gm-webhook-service.gm-operator.svc"},
		KeyRequest: &csr.KeyRequest{A: "rsa", S: 2048},
	}
	reqBytes, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}

	c := http.Client{Timeout: time.Second * 2}
	resp, err := getCFSSLResp(c, "newcert", fmt.Sprintf(`{"request":%s}`, string(reqBytes)))
	if err != nil {
		return nil, err
	}

	return []string{
		resp.Result.Cert,
		resp.Result.Key,
	}, nil
}

type cfsslResp struct {
	Result struct {
		Cert string `json:"certificate"`
		Key  string `json:"private_key"`
	} `json:"result"`
}

func getCFSSLResp(c http.Client, path, data string) (*cfsslResp, error) {
	url := fmt.Sprintf("http://127.0.0.1:8888/api/v1/cfssl/%s", path)
	req, err := http.NewRequest("POST", url, bytes.NewReader([]byte(data)))
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := &cfsslResp{}
	if err := json.Unmarshal(respBody, body); err != nil {
		return nil, err
	}
	return body, nil
}

// implements the cfssl/log.SysLogWriter interface
type cfsslLogger struct {
	logger logr.Logger
}

func (cl *cfsslLogger) Debug(msg string) {
	cl.logger.Info(msg, "level", "debug")
}

func (cl *cfsslLogger) Info(msg string) {
	cl.logger.Info(msg)
}

func (cl *cfsslLogger) Warning(msg string) {
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
