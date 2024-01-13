.. include:: /substitutions

Glossary
========

A helpful list of terms and definitions, potentially along with external resources to make it as easy as possible to understand the contents of this documentation.

.. glossary::

  |akm|
    Authentik Manager, this very project. A |declarative| |operator| to enable |iac| automated management of |auth| and |authorization| under |authentik|

    :see-also:

      https://gitlab.com/GeorgeRaven/authentik-manager

  authentication
    The process of verifying that someone is who they say they are by proving their identity. Not to be confused with |authorization| they are two very distinct but commonly related things. Authentication is to ensure something is not counterfit.

  |authentik|
    An open-source |auth| and |authorization| management platform for modern web applications. It provides a centralized system for managing user authentication and authorization for applications, making it easier for organizations to secure their applications and meet compliance requirements. Authentik provides a range of |auth| methods, including single sign-on (SSO) and multi-factor authentication (MFA), as well as support for third-party authentication providers. It also includes a flexible authorization system for controlling access to resources within applications. Overall, Authentik provides a comprehensive platform for managing authentication and authorization for modern web applications.

    :see-also:

      https://goauthentik.io/

  |authorization|
    The process of determining whether a given person / identity has permission to perform an action on a resource like adding new data. Not to be confused with |auth| they are two very distinct but commonly related things. Authorization does not know or care if the identity is real or fraudulent, it is more information which some other process can use to decide what to do.

  |controller|
    A |k8s| Controller is a component in the Kubernetes system that watches the state of the cluster and makes changes as necessary to ensure that the desired state is maintained. Controllers can be used to manage the state of any resource in a cluster, such as pods, services, or custom resources defined by a Custom Resource Definition (CRD). They continuously monitor the state of the cluster and take action to bring the actual state in line with the desired state, making sure that the cluster remains in a consistent and desired state. Controllers are a key component in the Kubernetes architecture, providing automatic management and self-healing capabilities to the cluster.

    :see-also:

      https://kubernetes.io/docs/concepts/architecture/controller/

  Kubernetes
    An open-source platform for automating the deployment, scaling, and management of containerized applications. It provides a declarative way to manage resources in a cluster and offers features such as service discovery, scaling, and self-healing for high availability. Kubernetes runs in a cloud-agnostic environment and supports a wide range of deployment options.

    :see-also:

      https://kubernetes.io

  |crd|
    A Custom Resource Definition (CRD) is a way to extend the Kubernetes API by defining custom resources. CRDs allow users to create their own custom resources and define their own API objects, which can then be used in the same way as built-in resources. CRDs provide a flexible and extensible way to manage application-specific configuration in a Kubernetes cluster.

    :see-also:

      https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/

  |declarative|
    A way to define and codify some state. In our contexts this is oftern related to infrastructure as code.

  |gitops|
    A software development and deployment approach that uses Git as the single source of truth for managing and deploying applications and infrastructure. Changes are made through pull requests and automatically deployed through continuous delivery pipelines, ensuring predictability, automation, and audibility. Infrastructure is managed as code and described in a declarative manner, enabling version control and collaboration.

    :see-also:

      https://about.gitlab.com/topics/gitops/

  |helm|
    A package manager for Kubernetes that makes it easier to manage, install, and upgrade complex, multi-tier applications in a cluster. Helm provides a way to define, package, and deploy applications as "charts," which are collections of files that describe the desired state of the application and its components. Helm charts make it easier to manage the installation and upgrades of complex applications, and can be used to share and reuse application configurations.

    :see-also:

      https://helm.sh/

  |iac|
    Infrastructure as Code is a method of managing infrastructure by declaring the desired state, and letting some tool reconcile it with the actual state of the infrastructure to ensure that it is in line with the desired state.

  |oidc|
    OpenID Connect is a very popular standard for authentication and authorization based on OAuth 2.0.

    :see-also:

      https://openid.net/developers/how-connect-works/

  |operator|
    A |k8s| |operator| is a software extension to |k8s| that automates the management of complex, stateful applications. Operators use |crd|\ s and |controller|\ s to manage the desired state of an application and its components, ensuring that the application is healthy, updated, and scalable. Operators simplify the deployment and management of applications on Kubernetes by automating common tasks and providing a declarative approach to managing the application's state.

    :see-also:

      https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
