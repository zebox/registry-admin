### Registry Admin
This  project allows manage repositories, images and users access  for self-hosted private docker registry with web UI. 
Main idea for implement this project to make high-level API for management user access to private registry 
and restrict their action (push/pull) for specific repositories (only with `token` auth) based 
on [official](https://docs.docker.com/registry/) private docker registry [image](https://hub.docker.com/_/registry). 
But someone need simple management to registry without split access to repositories. 
RegistryAdmin allows use either `password` or `token` authentication scheme for access management, 
depending on your task.  This application can be deployed with existed private registry for add access management 
UI tools to it.

Web user interface create with [React-Admin](https://marmelab.com/react-admin) framework and [MUI](https://mui.com/) components.

### Features
* Management users and access to registry
* Restrict access to repository per user action (`pull`/`push`, only for `token` auth scheme)
* List all repositories/images
* List all tags of image
* Display tag and image data
* Display image history
* Delete image tag
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


### Development Guidelines
- For local `frontend` development your should run **RegistryAdmin** server on port `80` and SSL disabled 
otherwise provider inside frontend can't get access to **RegistryAdmin** service. 
Also, you need use environment variable `RA_DEV_HOST=http://127.0.0.1:3000` for prevent `CORS` error in a browser.  