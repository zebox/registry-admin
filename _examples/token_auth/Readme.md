### Example with `token` authentication type

1. Grant permission for RegistryAdmin user inside container:

```bash
chown -R 1001:1001 {root-registry-admin-folder}
```

2. Set up environment `$REGISTRY_AUTH_TOKEN_REALM` for external hostname or IP address for `realm` URL in
   `docker-compose.yml` for `registry` service.
3. Change other options, if needed
4. Run services with docker-compose