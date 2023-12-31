.. include:: /substitutions

.. _section_akbp:

AkBlueprint
===========

|crd| for low-level |authentik| configuration. This resource will populate |authentik| with behaviours and integrations of your choosing at a very low level. This is one of the most involved and advanced |crd|\ s. This |crd| is usually the way other higher level |crd|\ s configure |authentik| to do things like forward authentication etc.

.. note::

  This resource should be placed in the same namespace (auth in this case) as the |operator| so that cluster permissions of the operator can be kept minimal.

.. |ak-fig| image:: /img/akbp.svg
  :width: 400
  :alt: AkBlueprint resource being used to populate postgresql

|ak-fig|

Currently only file-based blueprints are supported, direct-to-database blueprints are broadly implemented but a lot of quality of life is still missing like custom YAML tag support.

Spec
----

.. code-block:: yaml
   :caption: akblueprint-sample.yaml | A simple AkBlueprint CRD that changes the welcome message of authentik

    apiVersion: akm.goauthentik.io/v1alpha1
    kind: AkBlueprint
    metadata:
      labels:
        app.kubernetes.io/name: akblueprint
        app.kubernetes.io/instance: akblueprint-sample
        app.kubernetes.io/part-of: operator
        app.kubernetes.io/managed-by: kustomize
        app.kubernetes.io/created-by: operator
      name: akblueprint-sample
      namespace: auth
    spec:
      file: /blueprints/operator/blueprint-sample.yml
      blueprint: |
        version: 1
        metadata:
          labels:
            source: akm
          name: blueprint-sample
        entries:
        - model: authentik_flows.flow
          state: present
          identifiers:
            slug: akm-sample
          id: akm-flow
          attrs:
            denied_action: message_continue
            designation: stage_configuration
            name: default-oobe-setup
            title: Welcome to authentik!

See Also
--------

- Blueprints https://goauthentik.io/developer-docs/blueprints/
