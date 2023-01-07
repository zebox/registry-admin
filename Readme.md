<div align="center">
  <img src="assets/registry_admin_logo.svg" alt="logo" width="400"/>
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/zebox/registry-admin)](https://goreportcard.com/report/github.com/zebox/registry-admin)
[![Coverage Status](https://coveralls.io/repos/github/zebox/registry-admin/badge.svg?branch=master)](https://coveralls.io/github/zebox/registry-admin?branch=master)
[![Build Status](https://github.com/zebox/registry-admin/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/zebox/registry-admin/actions)
![Build Status](https://github.com/zebox/registry-admin/actions/workflows/ci-build.yml/badge.svg)

The RegistryAdmin is the UI tool that allows users to manage access to a private Docker registry.
It provides a web-based user interface for managing repositories, images, and user access, and allows users to
authenticate using either `password`. The main goal of the project is to provide a high-level API for managing user access
to a private registry, and to restrict user actions (such as push and pull) for specific repositories based on
the [official](https://docs.docker.com/registry/) private Docker registry [image](https://hub.docker.com/_/registry).
This can be useful for registry owners who want to have more control over their registry and who want to be able to manage access
to it more easily.

Web user interface created with [React-Admin](https://marmelab.com/react-admin) framework and [MUI](https://mui.com/)
components.
<div align="center">
  <img src="assets/demo_1.png" alt="demo"/>
</div>

### Features

* User management and access control for a private registry
* Restricted access to repository based on user actions (`pull`/`push`) for `token` authentication scheme)
* List all repositories and images
* List all tags for a specific image
* Displaying tag and image data
* Displaying the history of an image
* Deleting image tags
* Sharing anonymous access to specific repositories
* Sharing access to specific repositories with registered users only
* A built-in self-signed certificate builder
* Single binary distribution
* Distribution as a single binary or Docker container
* Automatic SSL termination using Let's Encrypt for access to the UI
* Logging options, including Apache Log Format and simplified stdout reports

---
RegistryAdmin is a tool that works in conjunction with a private Docker registry and uses the
registry's [V2 API](https://docs.docker.com/registry/spec/api/)
to communicate with it. It has the HTTP endpoint that is used to authenticate users using a token and to check their
access rights.
In order to use RegistryAdmin with a registry, the registry must be configured to support token-based authentication.
This allows users to be granted or restricted access to certain actions (such as `pull` or `push`) based on their
authentication token.

```yml
# in registry config file
...
auth:
  token:
  realm: https://{registry-admin-host}/api/v1/registry/auth
  service: container_registry_service_name
  issuer: registry_token_issuer_name
  rootcertbundle: /certs/cert.crt  # path to certificate bundle
```

You can use the `htpasswd` authentication scheme, but with this method you can only manage users and not restrict access
to repositories by specific user. This feature is only available when using token-based authentication.

To enhance the user experience with features such as sorting, searching, and autocompleting, RegistryAdmin has an
embedded
storage system that synchronizes with the data in the registry. This is necessary to avoid the limits of the search API
([catalog](https://docs.docker.com/registry/spec/api/#catalog)) exposed by the registry API, as the search function only
allows pagination with a cursor and does not support searching by repository entry. The app also includes an internal
garbage collector for check data consistency in the embedded storage.

To monitor changes in the registry, you should configure
registry [notification](https://docs.docker.com/registry/configuration/#notifications)
to be sent to the RegistryAdmin app.

```yml
# in registry config file
...
notifications:
  events:
    includereferences: true
  endpoints:
    - name: ra-listener
      disabled: false
      url: http://{registry-admin-host}/api/v1/registry/events
      headers:
      Authorization: [ Basic Y2xpZW50MDg6Y0t1cnNrQWdybzA4 ]
      timeout: 1s
      threshold: 5
      backoff: 3s
      ignoredmediatypes:
        - application/octet-stream
      ignore:
        mediatypes:
          - application/octet-stream
```

### Install
RegistryAdmin distributed as a small self-contained binary as well as a docker image. 
Both binary and image support multiple architectures and multiple operating systems,
including linux_x86_64, linux_arm64, linux_arm, macos_x86_64, macos_arm64, windows_x86_64 
and windows_arm. 

* for a binary distribution download the proper file in the release section
* docker container available on Docker Hub. I.e. docker pull zebox/registry-admin.

Latest stable version has :vX.Y.Z docker tag (with :latest alias) and the current master has :master tag.

### Configuration

#### 1. RegistryAdmin

To get started, you will need to set up the required parameters in a *docker-compose* file or using command line flags.
You can find various configuration examples in the [_examples](_examples) folder.

When you start RegistryAdmin as docker container you should set permission for application
folders (`certs`,`config`,`data`)
user with UID `1001

```bash
chown -R 1001:1001 {root-registry-admin-folder}
```

#### 1.1. Main setting

- `hostname` - define host name or IP address to include in the `AllowedOrigins` header which is used to check
  for `CORS` requests
- `port` - define the port that the application will use to listen for HTTP requests (the default is 80).
  Note: If you start the app as a Docker container, only ports `80` and `443` are exposed inside container.

- `store.type` - define storage type for store main data (users, accesses, repositories). Default (`embed`)

:warning: `Now implement embed storage type only`

- `store.admin_password` - overrides the default admin password when storage creating first (default password: `admin`)
- `store.embed.path ` - define pathname for embed storage file (default: `./data.db`)

#### 1.2.  RegistryAdmin settings (with token auth) - Recommended

- `registry.host` - define main host or IP address of private registry instance with protocol scheme prefix.

>`example: host: https://{registry-host}`
>

- `registry.port` - port of a private docker registry instance (default: `5000`)
- `registry.auth_type` - defines authenticate type `token` or `basic` (default: `token`).
- `issuer` - issuer name which checks inside registry, issuer name must be same at private docker registry and RegistryAdmin.
- `service` - service name which defined in registry settings, service name must be same at private docker registry and RegistryAdmin.

:exclamation: Keep a mind for `token` auth type required `certs` options must be defined.

- `registry.certs.path` - root directory where will be generated and stored certificates for token signing
- `registry.certs.key` - path to private key for token signing
- `registry.certs.public_key` - path to public key for verify token sign
- `registry.certs.ca` - path to certificate authority bundle
- `registry.certs.fqdns` - FQDN(s) required to add for registry certificate and checks at request from clients
- `registry.certs.ip` - an IP address will add to certificate extension field (SANs). If it omitted certificate error can be occurred.

:warning: If one fields is defined others should be defined too otherwise occurring an error

Certificates will be generated automatically if `registry.certs.path` is valid and directory is empty. If `certs` 
options isn't defined certificates will be created at a user home directory in sub folder `.registry-certs`:

```text
~/.registry-certs/
    registry_auth.key
    registry_auth.pub
    registry_auth_ca.crt
```
**Notice:** when self-signed certificates is used you should configure Docker Engine on a client host for work with ones.
```text
# https://docs.docker.com/config/daemon/
# /etc/docker/daemon.json (Linux)
# C:\ProgramData\docker\config\daemon.json (Windows)

{
 ...
 
  "insecure-registries": ["{registry-host}:{port}"],
  
 ...
}
```

Certificates generated for registry token also can be using for HTTP TLS/SSL. That certificate automatically add to trusted CA. 
But if you use another certificate for HTTPS access you should add it to trusted CA pool. For it use option  
`--registry.https-certs` for define path to a used certificate which using for TLS/SSL access. It's also required for certificates 
issued by Let's Encrypt. Also, you can define `--registry.https-insecure` option for skips check for the trusted certificate,
but in NOT RECOMMENDED.
```yml
# in a registry-admin config
registry:
  ...
  certs:
    ...
    https_cert: /{path-to-ssl}/cert.pem
    ...

```


#### 1.3. Private Docker Registry settings (with token auth) - Recommended

Supported registry V2 only. For use docker registry with token authentication you need configure it as a standalone
access control manager for resources hosted by other services which wish to authenticate and manage authorization
using a separate access control manager. For get more information about it, follow to the official
[documentations](https://docs.docker.com/registry/spec/auth/token/).

At first, you need define `auth` option for `token` auth and set specific `certificate` and `key` which generated with
the RegistryAdmin app. Token options must be the same as RegistryAdmin `Registry` defined
options (`issuer`,`service`,`cert_ca`). The RegistryAdmin app has public endpoint for authenticate users requests to
registry,
that must be used in `realm` registry option.

`https://{registry-admin-host}:{port}/api/v1/auth`

:exclamation: `realm` is option of *IP address* or *Hostname* RegistryAdmin instance that must accessible for docker
clients which
uses it for authenticate to private registry.

![Realm example in docker environment](assets/realm_example.png "Real example in docker environment")

```yml
auth:
  token:
    realm: http://{registry-admin-hostname}/api/v1/registry/auth
    service: container_registry
    issuer: registry_token_issuer
    rootcertbundle: /certs/cert.crt
```

For handle registry event and trigger repository task (such add new, update or delete repository entry) you should set
up
registry notification options:

`url` - http(s) url to RegistryAdmin host with events endpoint path.

`Authorization` any enabled and registered user and its password in the RegistryAdmin app, encoded in Base64.

```yml
notifications:
  events:
    includereferences: true
  endpoints:
    - name: ra-listener
      disabled: false
      url: http://registry-admin/api/v1/registry/events
      headers:
        Authorization: [Basic YWRtaW46c3VwZXItc2VjcmV0] # encoded in Base64 as 'admin:super-secret'
      timeout: 1s
      threshold: 5
      backoff: 3s
      ignoredmediatypes:
        - application/octet-stream
      ignore:
        mediatypes:
          - application/octet-stream
```

### RegistryAdmin settings (with basic auth, .htpasswd) - Not recommended

`basic` option using `.htpasswd` file and doesn't support restrict access to specific repositories. For use `basic`
authentication you required following options:

- `login` - username for access to docker registry
- `password` - password for access to docker registry

Docker registry reads `.htpasswd` file every time when authenticate call and doesn't require restart the registry
service
after user update or delete in RegistryAdmin

## Logging

By default, no request log generated. This can be turned on by setting `--logger.enabled`. The log (auto-rotated)
has [Apache Combined Log Format](http://httpd.apache.org/docs/2.2/logs.html#combined)

User can also turn stdout log on with `--logger.stdout`. It won't affect the file logging above but will output some
minimal info about processed requests, something like this:

```
127.0.0.1 - - [06/Dec/2022:18:36:34 +0300] "GET /auth/user HTTP/2.0" 200 159
127.0.0.1 - - [06/Dec/2022:18:36:34 +0300] "GET /api/v1/registry/auth HTTP/2.0" 200 198
```

### Additional Security (fail2ban)
When RegistryAdmin has access from Internet you should minimal set up security rules for prevent password brute force. 
The simplest way using `fail2ban` service  with access log file on a docker host.

1. Configure `access.log` for RegistryAdmin service
```yml
# in registry-admin config file
logger:
  enabled: true
  filename: /var/log/registry-admin/access.log # mount the directory to a docker host folder for get access for fail2ban
  max_size: 5M
  max_backups: 3

```
2. Create the `filter` with rule for the `registry-admin` service which handle `401` and `403` auth/z errors
```text
# /etc/fail2ban/filter.d/registry-admin.conf

[Definition]
failregex = ^<HOST> .+\" 40[1,3] \d+ .*$
```

3. Create `jail` with the `registry-admin` rule
```text
# in /etc/fail2ban/jail.local

[registry-admin]

enabled  = true
port     = http,https
filter   = registry-admin
logpath  = /{path-to-mounted-dir}/logs/access.log
findtime = 3600 # 1 hour or define your own
bantime  = 86400 # 24 hours or define your own
maxretry = 5
```


### Options

Each option can be provided in three forms: command line, environment key:value pair or config file (`json` or `yaml` formats).
Command line options have a long form only, like --hostname=localhost. The environment key (name) listed
for each option as a suffix, i.e. [$HOSTNAME].

```text

      --listen:                           listen on host:port (127.0.0.1:80/443 without) (default: *) [$RA_LISTEN]
      --hostname:                         Main hostname of service (default: localhost) [$RA_HOST_NAME]
      --port:                             Main web-service port. Default:80 (default: 80) [$RA_PORT]
      --config-file:                      Path to config file [$RA_CONFIG_FILE]
      --debug                             enable the debug mode [$RA_DEBUG]

registry:
      --registry.host:                    Main host or address to docker registry service [$RA_REGISTRY_HOST]
      --registry.port:                    Port which registry accept requests. Default:5000 (default: 5000) [$RA_REGISTRY_PORT]
      --registry.auth-type:[basic|token]  Type for auth to docker registry service. Available 'basic' and 'token'. Default 'token' (default: token) [$RA_REGISTRY_AUTH_TYPE]
      --registry.login:                   Username is a credential for access to registry service using basic auth type [$RA_REGISTRY_LOGIN]
      --registry.password:                Password is a credential for access to registry service using basic auth type [$RA_REGISTRY_PASSWORD]
      --registry.htpasswd:                Path to htpasswd file when basic auth type selected [$RA_REGISTRY_HTPASSWD]
      --registry.https-insecure           Set https connection to registry insecure [$RA_REGISTRY_HTTPS_INSECURE]
      --registry.service:                 A service name which defined in registry settings [$RA_REGISTRY_SERVICE]
      --registry.issuer:                  A token issuer name which defined in registry settings [$RA_REGISTRY_ISSUER]
      --registry.gc-interval:             Use for define custom time interval for garbage collector execute (minutes), default 1 hours [$RA_REGISTRY_GC_INTERVAL]

certs:
      --registry.certs.path:              A path to directory where will be stored new self-signed cert,keys and CA files, when 'token' auth type is used [$RA_REGISTRY_CERTS_CERT_PATH]
      --registry.certs.key:               A path where will be stored new self-signed private key file, when 'token' auth type is used [$RA_REGISTRY_CERTS_KEY_PATH]
      --registry.certs.public-key:        A path where will be stored new self-signed public key file, when 'token' auth type is used [$RA_REGISTRY_CERTS_PUBLIC_KEY_PATH]
      --registry.certs.ca-root:           A path where will be stored new CA bundles file, when 'token' auth type is used [$RA_REGISTRY_CERTS_CA_ROOT_PATH]
      --registry.certs.fqdn:              FQDN(s) for registry certificates [$RA_REGISTRY_CERTS_FQDN]
      --registry.certs.ip:                Address which appends to certificate SAN (Subject Alternative Name) [$RA_REGISTRY_CERTS_IP]
      --registry.https-certs:             A path to a HTTPS certificate used for TLS access to registry instance [$RA_REGISTRY_HTTPS_CERT]

auth:
      --auth.token-secret:                Main secret for auth token sign [$RA_AUTH_TOKEN_SECRET]
      --auth.jwt-issuer:                  Token issuer signature (default: zebox) [$RA_AUTH_ISSUER_NAME]
      --auth.jwt-ttl:                     Define JWT expired timeout (default: 1h) [$RA_AUTH_JWT_TTL]
      --auth.cookie-ttl:                  Define cookies expired timeout (default: 24h) [$RA_AUTH_COOKIE_TTL]

logger:
      --logger.stdout                     enable stdout logging [$RA_LOGGER_STDOUT]
      --logger.enabled                    enable access and error rotated logs [$RA_LOGGER_ENABLED]
      --logger.file:                      location of access log (default: access.log) [$RA_LOGGER_FILE]
      --logger.max-size:                  maximum size before it gets rotated (default: 10M) [$RA_LOGGER_SIZE]
      --logger.max-backups:               maximum number of old log files to retain (default: 10) [$RA_LOGGER_BACKUPS]

ssl:
      --ssl.type:[none|static|auto]       ssl (auto) support. Default is 'none' (default: none) [$RA_SSL_TYPE]
      --ssl.cert:                         path to cert.pem file [$RA_SSL_CERT]
      --ssl.key:                          path to key.pem file [$RA_SSL_KEY]
      --ssl.acme-location:                dir where certificates will be stored by autocert manager (default: ./acme) [$RA_SSL_ACME_LOCATION]
      --ssl.acme-email:                   admin email for certificate notifications [$RA_SSL_ACME_EMAIL]
      --ssl.port:                         Main web-service secure SSL port. Default:443 (default: 443) [$RA_SSL_PORT]
      --ssl.http-port:                    http port for redirect to https and acme challenge test (default: 80) [$RA_SSL_ACME_HTTP_PORT]
      --ssl.fqdn:                         FQDN(s) for ACME certificates [$RA_SSL_ACME_FQDN]

store:
      --store.type:[embed]                type of storage (default: embed) [$RA_STORE_DB_TYPE]
      --store.admin-password:             Define password for default admin user when storage create first (default: admin) [$RA_STORE_ADMIN_PASSWORD]

embed:
      --store.embed.path:                 Parent directory for the sqlite files (default: ./data.db) [$RA_STORE_EMBED_DB_PATH]

Help Options:
  -?                                     Show this help message
  -h, --help                              Show this help message
```

### Development Guidelines

- For local `frontend` development you should run RegistryAdmin with defined environment variable
  `RA_DEV_HOST=http://127.0.0.1:3000` for prevent `CORS` error in a browser. Also `.env.development` must contain valid
  development hostname of RegistryAdmin.
- Storage implement using `engine` [interface](app/store/engine/engine.go) and can be used for extends supported storage
  type
- `Embed` uses `SQLite` database and required `CGO` enabled
- For UI development required (if possible) use
  *react-admin* [guidelines](https://marmelab.com/react-admin/documentation.html)

### Status

The project is under active development and may have breaking changes till v1 is released. However, we are trying our
best not to break things unless there is a good reason.

### Credits

This project was inspired by projects and ideas of [Umputun](https://github.com/umputun)
