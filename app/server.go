package main

import (
	"context"
	"fmt"
	"github.com/zebox/registry-admin/app/registry"
	"github.com/zebox/registry-admin/app/store/engine/embedded"
	"io"
	"math"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/go-pkgz/auth"
	"github.com/go-pkgz/auth/avatar"
	"github.com/go-pkgz/auth/token"
	"github.com/pkg/errors"
	"github.com/zebox/registry-admin/app/server"
	"github.com/zebox/registry-admin/app/store/engine"
	"gopkg.in/natefinch/lumberjack.v2"

	log "github.com/go-pkgz/lgr"
)

func run() error {

	// setup logger for access requests
	accessLogger, err := createLoggerToFile()
	if err != nil {
		return errors.Wrap(err, "failed to setup logging to file, set logging to stdout")
	}

	defer func() {
		if logErr := accessLogger.Close(); logErr != nil {
			log.Printf("[WARN] can't close access log, %v", logErr)
		}
	}()

	tokenDuration, errTokenDuration := time.ParseDuration(opts.Auth.TokenDuration)
	if errTokenDuration != nil {
		return errTokenDuration
	}

	cookieDuration, errCookieDuration := time.ParseDuration(opts.Auth.TokenDuration)
	if errCookieDuration != nil {
		return errCookieDuration
	}

	sslConfig, sslErr := makeSSLConfig()
	if sslErr != nil {
		return fmt.Errorf("failed to make config of ssl server params: %w", sslErr)
	}

	registryService, errRegistry := createRegistryConnection(opts.Registry)
	if errRegistry != nil {
		return errRegistry
	}

	ctx, cancel := context.WithCancel(context.Background())
	dataStore, storeErr := makeDataStore(ctx, opts.Store)
	if storeErr != nil {
		cancel()
		return storeErr
	}

	srv := server.Server{
		Hostname:        opts.Auth.HostName,
		Listen:          opts.Listen,
		Port:            opts.Port,
		AccessLog:       accessLogger,
		L:               log.Default(),
		SSLConfig:       sslConfig,
		Storage:         dataStore,
		RegistryService: registryService,
	}

	authOptions := auth.Opts{
		SecretReader: token.SecretFunc(func(string) (string, error) { // secret key for JWT
			return opts.Auth.TokenSecret, nil
		}),
		ClaimsUpd:        token.ClaimsUpdFunc(srv.ClaimUpdateFn),
		TokenDuration:    tokenDuration,
		CookieDuration:   cookieDuration,
		Issuer:           opts.Auth.IssuerName,
		URL:              checkHostnameForURL(opts.Auth.HostName, opts.SSL.Type),
		BasicAuthChecker: srv.BasicAuthCheckerFn,
		AvatarStore:      avatar.NewNoOp(),
		SecureCookies:    true,
		DisableXSRF:      true,
		Validator:        &srv, // call Validate func for check token claims
		JWTQuery:         "jwt",
		Logger:           log.Default(),
	}

	authService := auth.NewService(authOptions)
	authService.AddDirectProvider("local", &srv)
	srv.Authenticator = authService

	go func() {
		if x := recover(); x != nil {
			log.Printf("[WARN] run time panic:\n%v", x)
			panic(x)
		}

		// catch signal and invoke graceful termination
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Printf("[WARN] interrupt signal")
		cancel()
	}()

	// shutdown server instance on context cancellation
	go func() {
		<-ctx.Done()
		log.Print("[INFO] shutdown initiated")
		srv.Shutdown()
	}()

	err = srv.Run(ctx)
	if err != nil && err == http.ErrServerClosed {
		log.Printf("[WARN] proxy server closed, %v", err) // nolint gocritic
	}
	return err
}

// checkHostnameForURL check hostname URL for valid format with specific scheme
func checkHostnameForURL(hostname, sslMode string) string {

	if !strings.HasPrefix(hostname, "http") && sslMode == "none" {
		return "http://" + hostname[:]
	}

	if !strings.HasPrefix(hostname, "http") && sslMode != "none" {
		return "https://" + hostname[:]
	}

	return hostname
}

// createRegistryConnection will prepare registry connection instance
func createRegistryConnection(opts RegistryGroup) (*registry.Registry, error) {

	var registrySettings registry.Settings

	if opts.Host == "" {
		return nil, errors.New("registry host undefined")
	}

	if opts.Port == 0 || opts.Port > 65535 {
		return nil, errors.New("wrong value of registry port")
	}

	// var re = regexp.MustCompile(`(?i)https?://`)
	// registrySettings.Host = re.ReplaceAllString(opts.Host, ``)
	registrySettings.Host = opts.Host
	registrySettings.Port = opts.Port
	registrySettings.Host = strings.TrimRight(registrySettings.Host, "/")

	// select registry auth type
	switch opts.AuthType {
	case "basic":
		registrySettings.AuthType = registry.Basic
		if opts.Htpasswd == "" {
			return nil, errors.New("htpasswd file path required for basic auth type")
		}
	case "self_token":
		registrySettings.Service = opts.Service
		registrySettings.Issuer = opts.Issuer
		registrySettings.AuthType = registry.SelfToken
	default:
		return nil, errors.Errorf("registry auth type '%s' not support", opts.AuthType)
	}

	if registrySettings.AuthType == registry.SelfToken {

		// paths to private, public keys and CA certificates for token generation if 'self_token' auth type defined
		registrySettings.CertificatesPaths.RootPath = opts.Certs.Path
		registrySettings.CertificatesPaths.KeyPath = opts.Certs.Key
		registrySettings.CertificatesPaths.PublicKeyPath = opts.Certs.PublicKey
		registrySettings.CertificatesPaths.CARootPath = opts.Certs.CARoot
	}

	return registry.NewRegistry(opts.Login, opts.Password, opts.Secret, registrySettings)
}

func sizeParse(inp string) (uint64, error) {
	if inp == "" {
		return 0, errors.New("empty value")
	}
	for i, sfx := range []string{"k", "m", "g", "t"} {
		if strings.HasSuffix(inp, strings.ToUpper(sfx)) || strings.HasSuffix(inp, strings.ToLower(sfx)) {
			val, err := strconv.Atoi(inp[:len(inp)-1])
			if err != nil {
				return 0, fmt.Errorf("can't parse %s: %w", inp, err)
			}
			return uint64(float64(val) * math.Pow(float64(1024), float64(i+1))), nil
		}
	}
	return strconv.ParseUint(inp, 10, 64)
}

// createLoggerToFile setup logger to file with rotation and backup
// forward to stdout if logger setup failed
func createLoggerToFile() (accessLog io.WriteCloser, err error) {
	if !opts.Logger.Enabled {
		return os.Stdout, nil
	}

	maxSize, perr := sizeParse(opts.Logger.MaxSize)
	if perr != nil {
		return os.Stdout, fmt.Errorf("can't parse logger MaxSize: %w", perr)
	}

	maxSize /= 1048576

	log.Printf("[INFO] logger enabled for %s, max size %dM", opts.Logger.FileName, maxSize)
	return &lumberjack.Logger{
		Filename:   opts.Logger.FileName,
		MaxSize:    int(maxSize), // in MB
		MaxBackups: opts.Logger.MaxBackups,
		Compress:   true,
		LocalTime:  true,
	}, nil
}

func makeDataStore(ctx context.Context, storeOpts StoreGroup) (iStore engine.Interface, err error) {
	log.Printf("[INFO] make data store, type=%s", storeOpts.Type)

	switch storeOpts.Type {
	case "embed":
		e := embedded.NewEmbedded(storeOpts.Embed.Path)
		err = e.Connect(ctx)
		if err != nil && !errors.Is(err, embedded.ErrTableAlreadyExist) {
			return nil, err
		}
		return e, nil
	default:
		return nil, fmt.Errorf("unsupported store type %s", storeOpts.Type)
	}
}

func redirectHTTPPort(port int) int {
	// don't set default if any ssl.http-port defined by user
	if port != 0 {
		return port
	}

	return 80
}

// fqdns cleans space suffixes and prefixes which can sneak in from docker compose
func fqdns(domains []string) (res []string) {
	for _, v := range domains {
		res = append(res, strings.TrimSpace(v))
	}
	return res
}

// makeSSLConfig setup SSL config for use in main service
func makeSSLConfig() (config server.SSLConfig, err error) {
	switch opts.SSL.Type {
	case "none":
		config.SSLMode = server.SSLNone
	case "static":
		if opts.SSL.Cert == "" {
			return config, errors.New("path to cert.pem is required")
		}
		if opts.SSL.Key == "" {
			return config, errors.New("path to key.pem is required")
		}
		config.SSLMode = server.SSLStatic
		config.Cert = opts.SSL.Cert
		config.Key = opts.SSL.Key
		config.Port = opts.SSL.Port
		config.RedirHTTPPort = redirectHTTPPort(opts.SSL.RedirHTTPPort)
	case "auto":
		config.SSLMode = server.SSLAuto
		config.ACMELocation = opts.SSL.ACMELocation
		config.ACMEEmail = opts.SSL.ACMEEmail
		config.FQDNs = fqdns(opts.SSL.FQDNs)
		config.Port = opts.SSL.Port
		config.RedirHTTPPort = redirectHTTPPort(opts.SSL.RedirHTTPPort)
	default:
		return config, fmt.Errorf("invalid value %q for SSL_TYPE, allowed values are: none, static or auto", opts.SSL.Type)
	}
	return config, err
}
