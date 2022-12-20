// Option is a main set of service option
// Some ideas and piece of code borrow from projects of Umputun (https://github.com/umputun)

package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// configReader implement different file read implementation (json, yml, toml etc.)
type configReader interface {
	ReadConfigFromFile(pathToFile string, opts *Options) error
}

// Options the main parameters for the service
type Options struct {
	Listen     string `long:"listen" env:"RA_LISTEN" default:"*" description:"listen on host:port (127.0.0.1:80/443 without)" json:"listen"`
	HostName   string `long:"hostname" env:"RA_HOST_NAME" default:"localhost" description:"Main hostname of service" json:"hostname"`
	Port       int    `long:"port" env:"RA_PORT" description:"Main web-service port. Default:80" default:"80" json:"port"`
	ConfigPath string `long:"config-file" env:"RA_CONFIG_FILE" description:"Path to config file"`

	Registry RegistryGroup `group:"registry" namespace:"registry" env-namespace:"RA_REGISTRY" json:"registry"`

	Auth struct {
		TokenSecret    string `long:"token-secret" env:"TOKEN_SECRET" description:"Main secret for auth token sign" json:"token_secret" yaml:"token_secret"`
		IssuerName     string `long:"jwt-issuer" env:"ISSUER_NAME" default:"zebox" description:"Token issuer signature" json:"issuer_name" yaml:"issuer_name"` //
		TokenDuration  string `long:"jwt-ttl" env:"JWT_TTL" default:"1h" description:"Define JWT expired timeout" json:"jwt_ttl" yaml:"jwt_ttl"`
		CookieDuration string `long:"cookie-ttl" env:"COOKIE_TTL" default:"24h" description:"Define cookies expired timeout" json:"cookie_ttl" yaml:"cookie_ttl"`
	} `group:"auth" namespace:"auth" env-namespace:"RA_AUTH" json:"auth"`

	Logger struct {
		StdOut     bool   `long:"stdout" env:"STDOUT" description:"enable stdout logging" json:"stdout" yaml:"stdout"`
		Enabled    bool   `long:"enabled" env:"ENABLED" description:"enable access and error rotated logs" json:"enabled"`
		FileName   string `long:"file" env:"FILE"  default:"access.log" description:"location of access log" json:"filename" yaml:"filename"`
		MaxSize    string `long:"max-size" env:"SIZE" default:"10M" description:"maximum size before it gets rotated" json:"max_size"  yaml:"max_size"`
		MaxBackups int    `long:"max-backups" env:"BACKUPS" default:"10" description:"maximum number of old log files to retain" json:"max_backups" yaml:"max_backups"`
	} `group:"logger" namespace:"logger" env-namespace:"RA_LOGGER"`

	SSL struct {
		Type          string   `long:"type" env:"TYPE" description:"ssl (auto) support. Default is 'none'" choice:"none" choice:"static" choice:"auto" default:"none" json:"type"` // nolint
		Cert          string   `long:"cert" env:"CERT" description:"path to cert.pem file" json:"cert"`
		Key           string   `long:"key" env:"KEY" description:"path to key.pem file" json:"key"`
		ACMELocation  string   `long:"acme-location" env:"ACME_LOCATION" description:"dir where certificates will be stored by autocert manager" default:"./acme" json:"acme_location" yaml:"acme_location"`
		ACMEEmail     string   `long:"acme-email" env:"ACME_EMAIL" description:"admin email for certificate notifications" json:"acme_email" yaml:"acme_email"`
		Port          int      `long:"port" env:"PORT" description:"Main web-service secure SSL port. Default:443" default:"443" json:"port"`
		RedirHTTPPort int      `long:"http-port" env:"ACME_HTTP_PORT" description:"http port for redirect to https and acme challenge test (default: 80)" json:"redir_http_port" yaml:"redir_http_port"`
		FQDNs         []string `long:"fqdn" env:"ACME_FQDN" env-delim:"," description:"FQDN(s) for ACME certificates" json:"acme_fqdns" yaml:"acme_fqdns"`
	} `group:"ssl" namespace:"ssl" env-namespace:"RA_SSL" json:"ssl"`

	Store StoreGroup `group:"store" namespace:"store" env-namespace:"RA_STORE" json:"store" yaml:"store"`
	Debug bool       `long:"debug" env:"RA_DEBUG" description:"enable the debug mode" json:"debug"`

	// implement interface for parse different types of config files
	configReader
}

// StoreGroup options which defined main storage instance
// Type implement as options for add support for different storage
type StoreGroup struct {
	Type          string `long:"type" env:"DB_TYPE" description:"type of storage" choice:"embed" default:"embed" json:"type"` // nolint
	AdminPassword string `long:"admin-password" env:"ADMIN_PASSWORD" description:"Define password for default admin user when storage create first" default:"admin" json:"admin_password" yaml:"admin_password"`
	Embed         struct {
		Path string `long:"path" env:"DB_PATH" default:"./data.db" description:"Parent directory for the sqlite files" json:"path" yaml:"path"`
	} `group:"embed" namespace:"embed" env-namespace:"EMBED" json:"embed" yaml:"embed"`
}

// RegistryGroup main setting for connection to private registry instance
type RegistryGroup struct {
	Host                     string `long:"host" env:"HOST" required:"true" description:"Main host or address to docker registry service" json:"host"`
	Port                     uint   `long:"port" env:"PORT" description:"Port which registry accept requests. Default:5000" default:"5000" json:"port"`
	AuthType                 string `long:"auth-type" env:"AUTH_TYPE" description:"Type for auth to docker registry service. Available 'basic' and 'token'. Default 'token'" choice:"basic" choice:"token" default:"token" json:"auth_type" yaml:"auth_type"`
	Login                    string `long:"login" env:"LOGIN" description:"Username is a credential for access to registry service using basic auth type" json:"login"`
	Password                 string `long:"password" env:"PASSWORD" description:"Password is a credential for access to registry service using basic auth type" json:"password"`
	Htpasswd                 string `long:"htpasswd" env:"HTPASSWD" description:"Path to htpasswd file when basic auth type selected" json:"htpasswd"`
	InsecureConnection       bool   `long:"https-insecure" env:"HTTPS_INSECURE" description:"Set https connection to registry insecure" json:"https_insecure"`
	Service                  string `long:"service" env:"SERVICE" description:"A service name which defined in registry settings" json:"service"`
	Issuer                   string `long:"issuer" env:"ISSUER" description:"A token issuer name which defined in registry settings" json:"issuer"`
	GarbageCollectorInterval int64  `long:"gc-interval" env:"GC_INTERVAL" description:"Use for define custom time interval for garbage collector call (in hour), default 1 hours" json:"gc_interval" yaml:"gc_interval"`
	Certs                    struct {
		Path      string   `long:"path" env:"CERT_PATH" description:"A path to directory where will be stored new self-signed cert,keys and CA files, when 'token' auth type is used" json:"path" yaml:"path"`
		Key       string   `long:"key" env:"KEY_PATH" description:"A path where will be stored new self-signed private key file, when 'token' auth type is used" json:"key"`
		PublicKey string   `long:"public-key" env:"PUBLIC_KEY_PATH" description:"A path where will be stored new self-signed public key file, when 'token' auth type is used" json:"public_key" yaml:"public_key"`
		CARoot    string   `long:"ca-root" env:"CA_ROOT_PATH" description:"A path where will be stored new CA bundles file, when 'token' auth type is used" json:"ca_root" yaml:"ca_root"`
		FQDNs     []string `long:"fqdn" env:"FQDN" env-delim:"," description:"FQDN(s) for registry certificates" json:"fqdns" yaml:"fqdns"`
		IP        string   `long:"ip" env:"IP" description:"Address which appends to certificate SAN (Subject Alternative Name)" json:"ip"`
	} `group:"certs" namespace:"certs" env-namespace:"CERTS" json:"certs"`
}

func parseArgs() (*Options, error) {
	var options Options
	_, errParse := flags.ParseArgs(&options, os.Args)

	// if config file undefined throw error when flag parse
	if options.ConfigPath == "" && errParse != nil {
		return nil, errors.Wrap(errParse, "failed to parse options failed")
	}

	if options.Port > 65535 || options.Port < 1 {
		return nil, errors.New("wrong port value")
	}

	// try read config from config file
	if options.ConfigPath != "" {
		ext := filepath.Ext(options.ConfigPath)
		switch ext {
		case ".json":
			options.configReader = new(jsonConfigParser)
			if errReadCfg := options.ReadConfigFromFile(options.ConfigPath, &options); errReadCfg != nil {
				return nil, errParse
			}
		case ".yml", ".yaml":
			options.configReader = new(yamlConfigParser)
			if errReadCfg := options.ReadConfigFromFile(options.ConfigPath, &options); errReadCfg != nil {
				return nil, errParse
			}
		default:
			return nil, errors.Errorf("config parser for %q not implemented", ext)
		}

	}

	if options.Auth.TokenSecret == "" {
		options.Auth.TokenSecret = generateRandomSecureToken(64)
		log.Print("No TokenSecret secret provided - generated random secret. To provide a TokenSecret, fill in " +
			"'token_secret' at 'auth' section in the configuration file, set the 'RA_AUTH_TOKEN_SECRET' environment variable " +
			"or use '--auth.token-secret' CLI flag.")
	}

	return &options, nil
}

// jsonConfigParser implementation of json file config parser
type jsonConfigParser struct{}

// ReadConfigFromFile the implement configReader interface method for json config file
func (j *jsonConfigParser) ReadConfigFromFile(pathToFile string, options *Options) error {
	data, errParse := os.ReadFile(filepath.Clean(pathToFile))
	if errParse != nil {
		return errors.Wrap(errParse, "failed to read json config file")
	}

	errParse = json.Unmarshal(data, options)
	if errParse != nil {
		return errors.Wrap(errParse, "failed to unmarshal json config data")
	}
	return nil
}

// yamlConfigParser implementation of yaml file config parser
type yamlConfigParser struct{}

// ReadConfigFromFile the implement configReader interface method for yaml config file
func (j *yamlConfigParser) ReadConfigFromFile(pathToFile string, options *Options) error {
	data, errParse := os.ReadFile(filepath.Clean(pathToFile))
	if errParse != nil {
		return errors.Wrap(errParse, "failed to read yaml config file")
	}
	errParse = yaml.Unmarshal(data, &options)
	if errParse != nil {
		return fmt.Errorf("failed to unmarshal yaml config data: %v", errParse)
	}
	return nil
}

// generateRandomSecureToken generates random secure token for sign JWT for authenticate.
// It's call if TokenSecret undefined in config parameters.
func generateRandomSecureToken(length int) string {
	b := make([]byte, length)
	if _, errRead := rand.Read(b); errRead != nil {
		return ""
	}
	return hex.EncodeToString(b)
}
