package version

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/gocode/gocodec"
	"github.com/go-redis/redis/v8"
)

type Version struct {
	cv cue.Value
}

func (v Version) Copy() Version {
	return Version{v.cv}
}

func (v Version) Values() Values {
	//lint:ignore SA1019 will update to Context in next Cue version
	codec := gocodec.New(&cue.Runtime{}, nil)
	var values Values
	codec.Encode(v.cv, &values)
	return values
}

func (v *Version) Apply(opts ...InstallOption) {
	for _, opt := range opts {
		opt(v)
	}
}

type InstallOption func(*Version)

var (
	spire         *cue.Value
	redisInternal *cue.Value
	redisExternal *cue.Value
)

// An InstallOption for injecting SPIRE configuration into Proxy values.
func SPIRE(v *Version) {
	if spire == nil {
		value, err := loadOption("spire.cue")
		if err != nil {
			return
		}
		spire = &value
	}
	v.cv = v.cv.Unify(*spire)
}

// An InstallOption for injecting configuration for an internal Redis deployment.
func InternalRedis(namespace string) InstallOption {
	return func(v *Version) {
		if redisInternal == nil {
			value, err := loadOption("redis_internal.cue")
			if err != nil {
				return
			}
			redisInternal = &value
		}
		b := make([]byte, 10)
		rand.Read(b)
		password := base64.URLEncoding.EncodeToString(b)
		injected, err := CueFromStrings(fmt.Sprintf(`
			namespace: "%s"
			password: "%s"
		`, namespace, password))
		if err != nil {
			logger.Error(err, "failed to inject apply option", "option", "internal Redis")
			return
		}
		v.cv = v.cv.Unify(*redisInternal).Unify(injected)
	}
}

// TODO: Instead of `url`, require host, port, password, dbs. No username option.
type ExternalRedisConfig struct {
	URL string `json:"url"`
	// +optional
	CertSecretName string `json:"cert_secret_name"`
}

// An InstallOption for injecting configuration for an external Redis server.
func ExternalRedis(cfg *ExternalRedisConfig) InstallOption {
	// TODO: In the Mesh validating webhook, ensure the user provided URL is parseable.
	// This actually might be OBE if we require the user to supply values separately rather than as a URL.
	redisOptions, _ := redis.ParseURL(cfg.URL)
	hostPort := redisOptions.Addr
	split := strings.Split(hostPort, ":")
	host, port := split[0], split[1]
	password := redisOptions.Password
	db := fmt.Sprintf("%d", redisOptions.DB)

	return func(v *Version) {
		if redisExternal == nil {
			value, err := loadOption("redis_external.cue")
			if err != nil {
				return
			}
			redisExternal = &value
		}
		injected, err := CueFromStrings(fmt.Sprintf(`
			host: "%s"
			port: "%s"
			password: "%s"
			db: "%s"
		`, host, port, password, db))
		if err != nil {
			logger.Error(err, "failed to inject apply option", "option", "internal Redis")
			return
		}
		v.cv = v.cv.Unify(*redisExternal).Unify(injected)
	}
}

func loadOption(fileName string) (cue.Value, error) {
	data, err := filesystem.ReadFile(fmt.Sprintf("options/%s", fileName))
	if err != nil {
		logger.Error(err, "failed to load option", "file", fileName)
		return cue.Value{}, err
	}
	value, err := CueFromStrings(string(data))
	if err != nil {
		logger.Error(err, "failed to parse option", "file", fileName)
	}
	return value, nil
}
