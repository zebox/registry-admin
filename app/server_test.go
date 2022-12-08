package main

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zebox/registry-admin/app/store/engine"
	"os"
	"strconv"
	"testing"
	"time"
)

func Test_redirHTTPPort(t *testing.T) {
	tbl := []struct {
		port int

		res int
	}{
		{0, 80},
		{0, 80},
		{1234, 1234},
		{1234, 1234},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.res, redirectHTTPPort(tt.port))
		})
	}
}

func Test_sizeParse(t *testing.T) {

	tbl := []struct {
		inp string
		res uint64
		err bool
	}{
		{"1000", 1000, false},
		{"0", 0, false},
		{"", 0, true},
		{"10K", 10240, false},
		{"1k", 1024, false},
		{"14m", 14 * 1024 * 1024, false},
		{"7G", 7 * 1024 * 1024 * 1024, false},
		{"170g", 170 * 1024 * 1024 * 1024, false},
		{"17T", 17 * 1024 * 1024 * 1024 * 1024, false},
		{"123aT", 0, true},
		{"123a", 0, true},
		{"123.45", 0, true},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, err := sizeParse(tt.inp)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.res, res)
		})
	}
}

func Test_checkHostnameForURL(t *testing.T) {
	tbl := []struct {
		origin  string
		result  string
		sslMode string
	}{
		{
			"127.0.0.1",
			"http://127.0.0.1",
			"none",
		},
		{
			"127.0.0.1",
			"https://127.0.0.1",
			"static",
		},
		{
			"http://127.0.0.1",
			"http://127.0.0.1",
			"none",
		},
		{
			"https://127.0.0.1",
			"https://127.0.0.1",
			"static",
		},
	}

	for i, tt := range tbl {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, tt.result, checkHostnameForURL(tt.origin, tt.sslMode))
		})
	}
}

func Test_makeDataStore(t *testing.T) {
	var sg = StoreGroup{
		Type: "embed",
		Embed: struct {
			Path string `long:"path" env:"DB_PATH" default:"./data.db" description:"Parent directory for the sqlite files" json:"path" yaml:"path"`
		}(struct {
			Path string `long:"path" env:"DB_PATH" default:"./data.db" description:"Parent directory for the sqlite files" json:"path"`
		}(struct {
			Path string `long:"path" env:"DB_PATH" default:"./data.db" description:"parent directory for the sqlite files" json:"path"`
		}(struct{ Path string }{Path: os.TempDir() + "/test_db"}))),
	}
	var (
		iStore       engine.Interface
		errNo, errIs error
	)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

	iStore, errNo = makeDataStore(ctx, sg)
	defer func() {
		cancel()
		assert.NoError(t, os.RemoveAll(os.TempDir()+"/test_db"))
	}()

	assert.NoError(t, errNo)
	assert.NotNil(t, iStore)
	assert.NoError(t, iStore.Close(ctx))

	sg.Type = "unknown"
	iStore, errIs = makeDataStore(ctx, sg)
	assert.Error(t, errIs)
	assert.Equal(t, iStore, nil)
}

func Test_createRegistryConnection(t *testing.T) {
	tmpDir, errDir := os.MkdirTemp("", "test_cert")
	require.NoError(t, errDir)

	defer func() {
		assert.NoError(t, os.RemoveAll(tmpDir))
	}()

	opts := Options{
		Registry: RegistryGroup{
			Host:     "http://localhost/",
			IP:       "127.0.0.1",
			Port:     5000,
			AuthType: "token",
			Login:    "test_login",
			Password: "test_password",
			Htpasswd: ".test_htpasswd",
			Certs: struct {
				Path      string `long:"path" env:"CERT_PATH" description:"A path to directory where will be stored new self-signed cert,keys and CA files, when 'token' auth type is used" json:"path" yaml:"path"`
				Key       string `long:"key" env:"KEY_PATH" description:"A path where will be stored new self-signed private key file, when 'token' auth type is used" json:"key"`
				PublicKey string `long:"public-key" env:"PUBLIC_KEY_PATH" description:"A path where will be stored new self-signed public key file, when 'token' auth type is used" json:"public_key" yaml:"public_key"`
				CARoot    string `long:"ca-root" env:"CA_ROOT_PATH" description:"A path where will be stored new CA bundles file, when 'token' auth type is used" json:"ca_root" yaml:"ca_root"`
			}(struct {
				Path      string
				Key       string
				PublicKey string
				CARoot    string
			}{Path: tmpDir + "/", Key: tmpDir + "/test.key", PublicKey: tmpDir + "/test.pub", CARoot: tmpDir + "/test.crt"}),
		},
	}

	rc, err := createRegistryConnection(opts.Registry)
	assert.NoError(t, err)
	assert.NotNil(t, rc)

	{
		opts.Registry.AuthType = "basic"
		rc, err = createRegistryConnection(opts.Registry)
		assert.NoError(t, err)
		assert.NotNil(t, rc)
	}

	{
		opts.Registry.Htpasswd = ""
		_, err = createRegistryConnection(opts.Registry)
		assert.Error(t, err)
	}

	// test for error
	opts.Registry.AuthType = "unknown"
	rc, err = createRegistryConnection(opts.Registry)
	assert.Error(t, err)
	assert.Nil(t, rc)

	opts.Registry.Host = "http://127.0.0.1:39999"
	rc, err = createRegistryConnection(opts.Registry)
	assert.Error(t, err)
	assert.Nil(t, rc)

	opts.Registry.Port = 0
	rc, err = createRegistryConnection(opts.Registry)
	assert.Error(t, err)
	assert.Nil(t, rc)

	opts.Registry.Host = ""
	rc, err = createRegistryConnection(opts.Registry)
	assert.Error(t, err)
	assert.Nil(t, rc)

}
