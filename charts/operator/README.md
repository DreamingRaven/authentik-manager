# authentik-manager

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 2022.11.3](https://img.shields.io/badge/AppVersion-2022.11.3-informational?style=flat-square)

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| file://../crds | authentik-manager-crds | 0.1.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| operator.clusterRole.enabled | bool | `true` |  |
| operator.clusterRole.generate | bool | `true` |  |
| operator.clusterRole.name | string | `"authentik-manager"` |  |
| operator.clusterRoleBinding.enabled | bool | `true` |  |
| operator.clusterRoleBinding.generate | bool | `true` |  |
| operator.clusterRoleBinding.name | string | `"authentik-manager"` |  |
| operator.deployment.env | list | `[]` |  |
| operator.deployment.image | string | `"registry.gitlab.com/georgeraven/authentik-manager:latest"` |  |
| operator.deployment.imagePullPolicy | string | `"Always"` |  |
| operator.deployment.name | string | `"authentik-manager"` |  |
| operator.deployment.replicas | int | `3` |  |
| operator.enabled | bool | `true` |  |
| operator.labels[0].key | string | `"type"` |  |
| operator.labels[0].value | string | `"auth"` |  |
| operator.labels[1].key | string | `"app"` |  |
| operator.labels[1].value | string | `"authentik-manager"` |  |
| operator.ports | list | `[]` |  |
| operator.role.enabled | bool | `true` |  |
| operator.role.generate | bool | `true` |  |
| operator.role.name | string | `"authentik-manager"` |  |
| operator.roleBinding.enabled | bool | `true` |  |
| operator.roleBinding.generate | bool | `true` |  |
| operator.roleBinding.name | string | `"authentik-manager"` |  |
| operator.serviceAccount.enabled | bool | `true` |  |
| operator.serviceAccount.generate | bool | `true` |  |
| operator.serviceAccount.name | string | `"authentik-manager"` |  |

