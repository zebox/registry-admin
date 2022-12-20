![logo](assets/registry_admin.svg)

The RegistryAdmin project is a tool that allows users to manage access to a private Docker registry. 
It provides a web-based user interface for managing repositories, images, and user access, and allows users to authenticate 
using either `password`. The main goal of the project is to provide a high-level API for managing user access 
to a private registry, and to restrict user actions (such as push and pull) for specific repositories based on 
the [official](https://docs.docker.com/registry/) private Docker registry [image](https://hub.docker.com/_/registry).
This can be useful for users who want to have more control over their registry and who want to be able to manage access 
to it more easily. 

Web user interface create with [React-Admin](https://marmelab.com/react-admin) framework and [MUI](https://mui.com/) components.

### Features
* Management users and access to registry
* Restrict access to repository per user action (`pull`/`push`, only for `token` auth scheme)
* List all repositories/images
* List all tags of image
* Display tag and image data
* Display image history
* Delete image tag
* Share anonymous access to specific repositories
* Share access to specific repositories for registered user only
* Built in self-signed certificate builder
* Single binary distribution
* Docker container distribution
* Automatic SSL termination (for access to UI) with Let's Encrypt
* Optional logging with both Apache Log Format, and simplified stdout reports

---

RegistryAdmin work in connection with the docker registry and uses registry
[V2 API](https://docs.docker.com/registry/spec/api/) for communicate with it. The app has 
the http endpoint which using for authenticate user by token and check user access right.  
For its registry required configure for token authenticate. Only token-based authentication 
allows you to restrict access for users by they action (`pull`/`push`).
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
You can use `htpasswd` authenticate scheme, but in this case you can manage users only, 
but not restrict access to repository by specific user. It's allowed for token-based auth only.

For adapt UI for UX features, such sort, search, autocomplete, RegistryAdmin has embedded storage
which synchronize with data of registry. It's require for avoid limit of search API 
([catalog](https://docs.docker.com/registry/spec/api/#catalog)) exposed by registry API 
(search allow pagination with cursor only, without search by text). The app also internal 
garbage collector for data in embedded storage.

For observe change in registry you should configure registry [notification](https://docs.docker.com/registry/configuration/#notifications) 
to RegistryAdmin app

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
At first, you need setup required parameters in compose file or using command line flags. Various configuration example 
you can find in the `_examples` folder.

#### 1.1. Main setting
- `hostname` - defines host name or ip address which includes to `AllowedOrigins` header and using for `CORS` requests check
- `port` - defines port which application uses for listening http requests (default `80`). **Notice:** if you start app as
docker container only ports `80` and `443` exposed to outside.

- `store.type` - define storage type for store main data (users, accesses, repositories). Default (`embed`)

:warning: `Now implement embed storage type only`

- `store.admin_password` - overrides the default admin password when storage creating first (default password: `admin`)
- `store.embed.path ` - defined path and name for embed storage file (default password: `./data.db`)

#### 1.2.  RegistryAdmin settings (with token auth) - Recommended

- `registry.host` - defines main host or ip address of private registry instance with protocol scheme prefix.
  It's hostname will be included to certificate extension field (`4.2.1.7  Subject Alternative Name`) if self-signed certificate defined.

>`example: host: https://{registry-host}`
> 
- `registry.port` - port of private docker registry instance (default `5000`)
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

**Notice**: Certificates generated for registry token also can be using for HTTP TLS/SSL.

#### 1.3.  Private Docker Registry settings (with token auth) - Recommended
Supported registry V2 only. For use docker registry with token authentication you need configure it as a standalone 
access control manager for resources hosted by other services which wish to authenticate and manage authorizations 
using a separate access control manager. For get more information about it, follow to the official 
[documentations](https://docs.docker.com/registry/spec/auth/token/). 

At first, you need define `auth` option for `token` auth and set specific `certificate` and `key` which generated with the 
RegistryAdmin app. Token options must be the same as RegistryAdmin `Registry` defined options (`issuer`,`service`,`cert_ca`).

:exclamation: `real` option it *IP address* or *Hostname* RegistryAdmin instance must accessible for docker clients which 
uses it for authenticate to private registry.
![Real example in docker environment](assets/realm_example.png "Real example in docker environment")

```yml
auth:
  token:
    realm: http://{registry-admin-hostname}/api/v1/registry/auth
    service: container_registry
    issuer: registry_token_issuer
    rootcertbundle: /certs/cert.crt
```
For handle registry event and trigger repository task (such add new, update or delete repository entry) you should setup
registry notification options:

`url` - http(s) url to RegistryAdmin host with events endpoint path.

`Authorization` any enabled and registered user in the RegistryAdmin app.

```yml
notifications:
  events:
    includereferences: true
  endpoints:
    - name: ra-listener
      disabled: false
      url: http://registry-admin/api/v1/registry/events
      headers:
        Authorization: [Basic YWRtaW46c3VwZXItc2VjcmV0]
      timeout: 1s
      threshold: 5
      backoff: 3s
      ignoredmediatypes:
        - application/octet-stream
      ignore:
        mediatypes:
          - application/octet-stream
```

### Registry settings (with basic auth, .htpasswd) - Not recommended
`basic` option using `.htpasswd` file and doesn't support restrict access to specific repositories and required restart 
a registry service each time when users updated . For use `basic`
you required following options:

- `login` - username for access to docker registry
- `password` - password for access to docker registry

Docker registry reads `.htpasswd` file every time when authenticate call and doesn't required restart registry service 
after user update or delete in RegistryAdmin 
## Logging

By default, no request log generated. This can be turned on by setting `--logger.enabled`. The log (auto-rotated) has [Apache Combined Log Format](http://httpd.apache.org/docs/2.2/logs.html#combined)

User can also turn stdout log on with `--logger.stdout`. It won't affect the file logging above but will output some minimal info about processed requests, something like this:

```
127.0.0.1 - - [06/Dec/2022:18:36:34 +0300] "GET /auth/user HTTP/2.0" 200 159
127.0.0.1 - - [06/Dec/2022:18:36:34 +0300] "GET /api/v1/registry/auth HTTP/2.0" 200 198
```

### Options

Each option can be provided in three forms: command line, environment key:value pair or config file (`json` or `yaml` formats).
Command line options have a long form only, like --hostname=localhost. The environment key (name) listed
for each option as a suffix, i.e. [$HOSTNAME].

```text

      /listen:                           listen on host:port (127.0.0.1:80/443 without) (default: *) [$RA_LISTEN]
      /hostname:                         Main hostname of service (default: localhost) [$RA_HOST_NAME]
      /port:                             Main web-service port. Default:80 (default: 80) [$RA_PORT]
      /config-file:                      Path to config file [$RA_CONFIG_FILE]
      /debug                             enable the debug mode [$RA_DEBUG]

registry:
      /registry.host:                    Main host or address to docker registry service [$RA_REGISTRY_HOST]
      /registry.port:                    Port which registry accept requests. Default:5000 (default: 5000) [$RA_REGISTRY_PORT]
      /registry.auth-type:[basic|token]  Type for auth to docker registry service. Available 'basic' and 'token'. Default 'token' (default: token) [$RA_REGISTRY_AUTH_TYPE]
      /registry.login:                   Username is a credential for access to registry service using basic auth type [$RA_REGISTRY_LOGIN]
      /registry.password:                Password is a credential for access to registry service using basic auth type [$RA_REGISTRY_PASSWORD]
      /registry.htpasswd:                Path to htpasswd file when basic auth type selected [$RA_REGISTRY_HTPASSWD]
      /registry.https-insecure           Set https connection to registry insecure [$RA_REGISTRY_HTTPS_INSECURE]
      /registry.service:                 A service name which defined in registry settings [$RA_REGISTRY_SERVICE]
      /registry.issuer:                  A token issuer name which defined in registry settings [$RA_REGISTRY_ISSUER]
      /registry.gc-interval:             Use for define custom time interval for garbage collector call (in hour), default 1 hours [$RA_REGISTRY_GC_INTERVAL]

certs:
      /registry.certs.path:              A path to directory where will be stored new self-signed cert,keys and CA files, when 'token' auth type is used [$RA_REGISTRY_CERTS_CERT_PATH]
      /registry.certs.key:               A path where will be stored new self-signed private key file, when 'token' auth type is used [$RA_REGISTRY_CERTS_KEY_PATH]
      /registry.certs.public-key:        A path where will be stored new self-signed public key file, when 'token' auth type is used [$RA_REGISTRY_CERTS_PUBLIC_KEY_PATH]
      /registry.certs.ca-root:           A path where will be stored new CA bundles file, when 'token' auth type is used [$RA_REGISTRY_CERTS_CA_ROOT_PATH]
      /registry.certs.fqdn:              FQDN(s) for registry certificates [$RA_REGISTRY_CERTS_FQDN]
      /registry.certs.ip:                Address which appends to certificate SAN (Subject Alternative Name) [$RA_REGISTRY_CERTS_IP]

auth:
      /auth.token-secret:                Main secret for auth token sign [$RA_AUTH_TOKEN_SECRET]
      /auth.jwt-issuer:                  Token issuer signature (default: zebox) [$RA_AUTH_ISSUER_NAME]
      /auth.jwt-ttl:                     Define JWT expired timeout (default: 1h) [$RA_AUTH_JWT_TTL]
      /auth.cookie-ttl:                  Define cookies expired timeout (default: 24h) [$RA_AUTH_COOKIE_TTL]

logger:
      /logger.stdout                     enable stdout logging [$RA_LOGGER_STDOUT]
      /logger.enabled                    enable access and error rotated logs [$RA_LOGGER_ENABLED]
      /logger.file:                      location of access log (default: access.log) [$RA_LOGGER_FILE]
      /logger.max-size:                  maximum size before it gets rotated (default: 10M) [$RA_LOGGER_SIZE]
      /logger.max-backups:               maximum number of old log files to retain (default: 10) [$RA_LOGGER_BACKUPS]

ssl:
      /ssl.type:[none|static|auto]       ssl (auto) support. Default is 'none' (default: none) [$RA_SSL_TYPE]
      /ssl.cert:                         path to cert.pem file [$RA_SSL_CERT]
      /ssl.key:                          path to key.pem file [$RA_SSL_KEY]
      /ssl.acme-location:                dir where certificates will be stored by autocert manager (default: ./acme) [$RA_SSL_ACME_LOCATION]
      /ssl.acme-email:                   admin email for certificate notifications [$RA_SSL_ACME_EMAIL]
      /ssl.port:                         Main web-service secure SSL port. Default:443 (default: 443) [$RA_SSL_PORT]
      /ssl.http-port:                    http port for redirect to https and acme challenge test (default: 80) [$RA_SSL_ACME_HTTP_PORT]
      /ssl.fqdn:                         FQDN(s) for ACME certificates [$RA_SSL_ACME_FQDN]

store:
      /store.type:[embed]                type of storage (default: embed) [$RA_STORE_DB_TYPE]
      /store.admin-password:             Define password for default admin user when storage create first (default: admin) [$RA_STORE_ADMIN_PASSWORD]

embed:
      /store.embed.path:                 Parent directory for the sqlite files (default: ./data.db) [$RA_STORE_EMBED_DB_PATH]

Help Options:
  /?                                     Show this help message
  /h, /help                              Show this help message
```

### Development Guidelines
- For local `frontend` development you should run **RegistryAdmin** with defined environment variable 
`RA_DEV_HOST=http://127.0.0.1:3000` for prevent `CORS` error in a browser. Also `.env.development` must contain valid 
hostname to RegistryAdmin development host.   
- Storage implement using `engine` [interface](app/store/engine/engine.go) and can be used for extends supported storage type
- `Embed` uses `SQLite` database and required `CGO` enabled
- For UI development required (if possible) use *react-admin* [guidelines](https://marmelab.com/react-admin/documentation.html)

### Status
The project is under active development and may have breaking changes till v1 is released. However, we are trying our 
best not to break things unless there is a good reason. 