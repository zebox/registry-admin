// Option is a main set of service option
// Some ideas and piece of code borrow from projects of Umputun (https://github.com/umputun)

package main

import (
	"encoding/json"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// configReader implement different file read implementation (json, yml, toml etc.)
type configReader interface {
	ReadConfigFromFile(pathToFile string, opts *Options) error
}

// Options the main parameters for the service
type Options struct {
	Version      string
	Listen       string `short:"l" long:"listen" env:"LISTEN" description:"listen on host:port (127.0.0.1:80/443 without)" json:"listen"`
	Port         int    `short:"p" long:"port" env:"PORT" description:"Main web-service port. Default:80" default:"80" json:"port"`
	ConfigPath   string `short:"f" long:"config-file" env:"CONFIG_FILE" description:"Path to config file"`
	RegistryHost string `long:"registry-host" env:"REGISTRY_HOST" required:"true" description:"Main host or address to docker registry service" json:"registry_host"`
	Auth         struct {
		TokenSecret    string `long:"token-secret" env:"AUTH_TOKEN_SECRET" required:"true" description:"Main secret for token" json:"token_secret" `
		HostName       string `long:"hostname" env:"AUTH_HOST_NAME" default:"localhost" description:"Main hostname of service" json:"host_name"`
		IssuerName     string `long:"jwt-issuer" env:"AUTH_ISSUER_NAME" required:"true" default:"zebox" description:"Token issuer signature" json:"issuer_name"`
		TokenDuration  string `long:"jwt-ttl" env:"AUTH_JWT_TTL" default:"1h" description:"Define JWT expired timeout" json:"jwt_ttl"`
		CookieDuration string `long:"cookie-ttl" env:"AUTH_COOKIE_TTL" default:"10d" description:"Define cookies expired timeout" json:"cookie_ttl"`
	} `group:"auth" namespace:"auth" env-namespace:"AUTH" json:"authenticate"`

	Logger struct {
		StdOut     bool   `long:"stdout" env:"LOGGER_STDOUT" description:"enable stdout logging" json:"stdout"`
		Enabled    bool   `long:"enabled" env:"LOGGER_ENABLED" description:"enable access and error rotated logs" json:"enabled"`
		FileName   string `long:"file" env:"LOGGER_FILE"  default:"access.log" description:"location of access log" json:"file_name"`
		MaxSize    string `long:"max-size" env:"LOGGER_MAX_SIZE" default:"100M" description:"maximum size before it gets rotated" json:"max_size"`
		MaxBackups int    `long:"max-backups" env:"LOGGER_MAX_BACKUPS" default:"10" description:"maximum number of old log files to retain" json:"max_backups"`
	} `group:"logger" namespace:"logger" env-namespace:"LOGGER"`

	SSL struct {
		Type          string   `long:"type" env:"SSL_TYPE" description:"ssl (auto) support. Default is 'none'" choice:"none" choice:"static" choice:"auto" default:"none" json:"type"` // nolint
		Cert          string   `long:"cert" env:"SSL_CERT" description:"path to cert.pem file" json:"cert"`
		Key           string   `long:"key" env:"SSL_KEY" description:"path to key.pem file" json:"key"`
		ACMELocation  string   `long:"acme-location" env:"ACME_LOCATION" description:"dir where certificates will be stored by autocert manager" default:"./acme" json:"acme_location"`
		ACMEEmail     string   `long:"acme-email" env:"ACME_EMAIL" description:"admin email for certificate notifications" json:"acme_email"`
		Port          int      `long:"port" env:"SSL_PORT" description:"Main web-service secure SSL port. Default:443" default:"443" json:"port"`
		RedirHTTPPort int      `long:"http-port" env:"ACME_HTTP_PORT" description:"http port for redirect to https and acme challenge test (default: 80)" json:"redir_http_port"`
		FQDNs         []string `long:"fqdn" env:"ACME_FQDN" env-delim:"," description:"FQDN(s) for ACME certificates" json:"acme_fqdns"`
	} `group:"ssl" namespace:"ssl" env-namespace:"SSL" json:"ssl"`

	Store StoreGroup `group:"store" namespace:"store" env-namespace:"STORE" json:"store"`
	Debug bool       `long:"debug" env:"DEBUG" description:"debug mode" json:"debug"`

	configReader
}

// StoreGroup options which defined main storage instance
// Type implement as options for add support for different storage
type StoreGroup struct {
	Type  string `long:"type" env:"DB_TYPE" description:"type of storage" choice:"embed" default:"embed" json:"type"` // nolint
	Embed struct {
		Path string `long:"path" env:"DB_PATH" default:"./data.db" description:"parent directory for the sqlite files" json:"path"`
	} `group:"embed" namespace:"embed" env-namespace:"embed" json:"embed"`
}

func parseArgs() (*Options, error) {
	var options Options
	_, err := flags.ParseArgs(&options, os.Args)

	// if config file undefined throw error when flag parse
	if options.ConfigPath == "" && err != nil {
		return nil, errors.Wrap(err, "failed to parse options failed")
	}

	if options.Port > 65535 || options.Port < 1 {
		return nil, errors.New("wrong port value")
	}

	// try read config from config file
	if options.ConfigPath != "" {
		ext := filepath.Ext(options.ConfigPath)
		switch ext {
		case "json":
			options.configReader = new(jsonConfigParser)
			if err := options.ReadConfigFromFile(options.ConfigPath, &options); err != nil {
				return nil, err
			}
		default:
			return nil, errors.Errorf("config parser for '%s' not implemented", ext)
		}

	}
	return &options, nil
}

// jsonConfigParser implementation of json file config parser
type jsonConfigParser struct{}

// ReadConfigFromFile the implement configReader interface method
func (j *jsonConfigParser) ReadConfigFromFile(pathToFile string, options *Options) error {
	data, err := ioutil.ReadFile(pathToFile)
	if err != nil {
		return errors.Wrap(err, "failed to read config file")
	}

	err = json.Unmarshal(data, options)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal config data")
	}
	return nil
}
