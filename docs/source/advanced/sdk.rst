.. include:: /substitutions

.. _section_sdk:

Operator-SDK
============

.. note::

   You won't need this for any of the works in this repository. We added this as a helpful reminder for ourselves and others, in particular documenting how we arrived at the current artifacts.

We use a hybrid helm-go operator to allow us to simplify mass deployment but also enable us to deeply automate beyond what helm can do. Thus the best of both worlds.

.. code-block:: bash
   :caption: Hybrid operator-SDK initialisation

   operator-sdk init --plugins=hybrid.helm.sdk.operatorframework.io --project-version="3" --domain goauthentik.io --repo=gitlab.com/GeorgeRaven/authentik-manager

.. code-block:: bash
   :caption: Go based (Ak CRD) API resource

   operator-sdk create api --group=akm --version v1alpha1 --kind Ak --resource --controller --plugins=go/v3

See Also
--------

- Hybrid operator SDK initialisation https://docs.openshift.com/container-platform/4.10/operators/operator_sdk/helm/osdk-hybrid-helm.html
