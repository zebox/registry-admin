### Registry Admin

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
you can find in the `example` folder.

**1. Application setting**
- `hostname` - defines host name or ip address which includes to `AllowedOrigins` header and using for `CORS` requests check
- `port` - defines port which application uses for listening http requests (default `80`)
- `store.type` - define storage type for store main data (users, accesses, repositories). Default (`embed`)

[^note]: `Now implement only embed storage type`

- `store.admin_password` - overrides the default admin password when storage creating first (default password: `admin`)
- `store.embed.path ` - defined path and name for embed storage file (default password: `./data.db`)

**2. Registry settings**

- `registry.host` - defines main host or ip address of private registry instance with protocol scheme prefix.
  It's hostname will be included to certificate extension field (`4.2.1.7  Subject Alternative Name`) if self-signed certificate defined

[^note]: `host: https://{registry-host}`
- `registry.ip` - need for included to certificate extension field only. If it omitted certificate error can be occurred.
- `registry.port` - port of private docker registry instance (default `5000`)
- `registry.auth_type` - defines authenticate type `token` or `basic` (default: `token`).

`basic` option using `.htpasswd` file and doesn't support restrict access to specific repositories. For use `basic`
you required following options:
- `login` - username for access to docker registry
- `password` - password for access to docker registry

[^note]: `Keep a mind for `token` auth type required `certs` options must be defined.`
- `registry.certs.path` - root directory where will be generated and stored certificates for token signing
- `registry.certs.key` - path to private key for token signing
- `registry.certs.public_key` - path to public key for verify token sign
- `registry.certs.ca` - path to certificate authority bundle 




 
### Development Guidelines
- For local `frontend` development your should run **RegistryAdmin** server on port `80` and SSL disabled 
otherwise provider inside frontend can't get access to **RegistryAdmin** service. 
Also, you need use environment variable `RA_DEV_HOST=http://127.0.0.1:3000` for prevent `CORS` error in a browser.  