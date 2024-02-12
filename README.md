# Authentik Manager

[![docs](https://gitlab.com/GeorgeRaven/authentik-manager/badges/master/pipeline.svg)](https://georgeraven.gitlab.io/authentik-manager/)
[![GitLab](https://img.shields.io/gitlab/v/tag/41806964?color=teal&label=AKM&sort=semver&style=for-the-badge)](https://gitlab.com/GeorgeRaven/authentik-manager)

Authentik-Manager (AKM) is a hybrid operator that deploys, configures, and manages the life-cycle of authentik, declaratively. This operator is primarily geared towards GitOps and enabling authentik to be consistently, reproducibly, and collaboratively managed.

This work is still under heavy development, but please submit an issue if you do try it out, and let us know if there are any problems.

## Documentation Versions

| Version | Docs                                             |
|---------|--------------------------------------------------|
| master  | https://georgeraven.gitlab.io/authentik-manager/ |


## At a Glimpse

```mermaid

flowchart TD

    subgraph kubeNs[kube-system]
        api[api]
        %% etcd[etcd]
        %% dns[dns]
        %% cm[controller-manager]
        %% proxy[proxy]
    end

    subgraph akmNs[auth]
        subgraph aknp[akm network policy]
            akm[authentik-manager\nleader]
            akm2[authentik-manager\npassive]
            akm-.-|high availability|akm2
            akm-->|watch\nresources|api
            akm-->|reconcile\ncrds|api
            akm-->|leader\nelection|api
            akm2-->|leader\nelection|api
        end
        subgraph netpol[ak network policy]
            akmpod1[authentik-server]
            akmpod2[authentik-worker]
            akmpod1-->akmredis[redis]
            akmpod2-->akmredis[redis]
            akmsql[postgresql]
            akmpod1-->akmsql
            akmpod2-->akmsql
            pgadmin-->akmsql

        end
        ak[ak crd]
        akbpsec[configmap]
        ak-.->|produces|netpol
        akm-.-|reconcile|ak
        akbp-.->|produces|akbpsec
        akbpsec---|mount|akmpod2
        akbp[akblueprint]
        akm-.-|reconcile|akbp

        oidcProBp[akblueprint oidc provider]
        oidcProBpCm[configmap]
        oidcProBp-.->|produces|oidcProBpCm
        akm-.-|reconcile|oidcProBp
        oidcProBpCm---|mount|akmpod2

        oidcAppBp[akblueprint oidc application]
        oidcAppBpCm[configmap]
        oidcAppBp-.->|produces|oidcAppBpCm
        akm-.-|reconcile|oidcAppBp
        oidcAppBpCm---|mount|akmpod2

        akmpod1---|mount|akmconfig[configmap]
        akmpod1---|env|akmsecret[secret]

    end

    akm-.->|reconcile|oidc
    oidc-.->|produces|oidcAppBp
    oidc-.->|produces|oidcProBp

    subgraph app[some application]

        oidc[oidc crd]
        appoidc[oidc-app]
        oidcsec[secret]
        oidconfig[configmap]
        appoidc---|env|oidcsec
        appoidc---|env|oidcconfig
        oidc -.->|produces| oidcsec
        oidc -.->|produces| oidcconfig
    end

```
