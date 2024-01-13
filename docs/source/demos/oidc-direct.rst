.. include:: /substitutions

.. _section_oidc_direct_demo:

OIDC Direct Demo
================

.. note::

   This demo is a work in progress, and is not yet complete.

This demo will show you how to use |akm| to automatically configure |authentik| as an |oidc| provider. This demo also shows some basic Go implementation to connect to the provider and display a webpage which will show the user their own |oidc| information the client found.

Dependencies of this demo, please install these before starting:

- minikube (for creating a temporary |k8s| cluster)
- podman (a friendly interface for your containers)
- kubectl (used to interact with the cluster)
- |helm| (used to install |akm| and other helm charts)
- make (automation for building and running various local tasks)
- git (git client for cloning the repo)
- yq (yaml query tool)
- tr (translate characters)

If you just want to run the demo, you can run the following to download the source code, change directory and run the full demo:

.. code-block:: bash

   git clone https://gitlab.com/GeorgeRaven/authentik-manager && cd authentik-manager && make demo

The Plan
--------

First we need to plan what we want to do:

-   Create a |k8s| cluster
-   Deploy |akm| and configure it
-   Deploy an |oidc| application and configure it
-   Login using |oidc|

While this is all automated, we want to walk you through it so you can see how it works. This also enables you to potentially debug, and fix issues where and when they may occur.

Create a K8s cluster
--------------------

.. code-block:: bash

   minikube delete
   minikube start --cni calico --driver=podman --kubernetes-version=MINIKUBE_KUBE_VERSION

Replace MINIKUBE_KUBE_VERSION with the version of |k8s| you want to use.

This will create a temporary |k8s| cluster, and a container network called calico.

See Also
--------

- go-oidc https://github.com/coreos/go-oidc
- authentik oauth2 provider https://goauthentik.io/integrations/sources/oauth/
