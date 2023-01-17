This is example with self-signed certificate that require specific configure option for override starting command
for a `registry` container in the `docker-compose.yml`.

`command: ["/bin/sh", "-c", "cp /certs/cert.crt /usr/local/share/ca-certificates && /usr/sbin/update-ca-certificates; registry serve /etc/docker/registry/config.yml"]`

If using a Let's Encrypt certificate obtained with `HTTP-01 challenge` you should define `fullchain.pem` file
for HTTPS SSL/TLS connection. If using ACME proto for obtain LE certs you must define ACME cache file for both config
`registy` and `registry-admin`.

:exclamation: Defined in this example host `registry.local` must be changed for real host in your network environment.