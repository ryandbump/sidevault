# sidevault

A small container for handling [Hashicorp Vault](https://www.vaultproject.io/)
token management in Kubernetes. The `sidevault` application has two use-cases
that are often used in conjunction with each other.

1. Perform [Vault Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes.html#authentication)
and store the resulting Vault token in a configurable shared volume. This is
most commonly done using an [init container](https://kubernetes.io/docs/concepts/workloads/pods/init-containers/)
1. Monitor an existing Vault token and attempt to renew once the remaining TTL
reaches a threshold. This is most commonly done using a sidecar container.

## Configuration

### Vault

`sidevault` uses the official [Vault Go Client](https://github.com/hashicorp/vault/tree/master/api)
and expects you to handle any Vault configuration through the environment. 

See the [Vault documentation](https://github.com/hashicorp/vault/tree/master/api)
on environment variables for more information.

### sidevault

All `sidevault` configuration can be provided through environment variables or
CLI flags. Flags always take precedence over the environment variables.

#### Global

- `TOKEN_PATH`, `--token-path <>`
    - file system path to the Vault token
    - default: `/var/run/secrets/vaultproject.io/.vault-token`
- `ACCESSOR_PATH`, `--accessor-path <>`
    - file system path to the Vault token accessor
    - default: `/var/run/secrets/vaultproject.io/.vault-accessor`

#### Auth

- `ROLE`, `--role <>`
    - role to use for Vault authentication
    - **required**
- `MOUNT_PATH`, `--mount-path <>`
    - mount path for the Kubernetes authentication backend
    - default: `kubernetes`
- `SA_TOKEN_PATH`, `--sa-token-path <>`
    - file system path to the Kubernetes ServiceAccount token
    - default: `/var/run/secrets/kubernetes.io/serviceaccount/token`

#### Renew

- `FREQUENCY`, `--frequency <>`
    - delay between token ttl checks, in seconds
    - default: `30`
- `LEASE`, `--lease <>`
    - request lease increment when renewing, in seconds
    - default: token creation TTL

## Example

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: backend
spec:
  terminationGracePeriodSeconds: 0
  containers:
    - name: backend
      ...
      volumeMounts:
        - mountPath: /etc/vault
          name: vault-token
    - name: sidevault-renew
      image: ryandbump/sidevault:0.0.1
      args:
        - renew
      volumeMounts:
        - name: vault-token
          mountPath: /var/run/secrets/vaultproject.io
      env:
        - name: VAULT_ADDR
          value: http://vault.default:8200
  initContainers:
    - name: sidevault-auth
      image: ryandbump/sidevault:0.0.1
      args:
        - auth
      volumeMounts:
        - name: vault-token
          mountPath: /var/run/secrets/vaultproject.io
      env:
        - name: ROLE
          value: backend
        - name: VAULT_ADDR
          value: http://vault.default:8200
  volumes:
    - name: vault-token
      emptyDir:
        medium: Memory
```
