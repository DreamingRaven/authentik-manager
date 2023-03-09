.. include:: /substitutions

.. _section_ak:

Ak
==

|crd| for |authentik| deployment, and configuration. This resource dictates where and how an |authentik| stack should be deployed.

.. note::

  This resource should be placed in the same namespace (auth in this case) as the |operator| so that cluster permissions of the operator can be kept minimal.

.. |ak-fig| image:: /img/ak.svg
  :width: 400
  :alt: Ak CRD resource control

|ak-fig|

Spec
----

This will create a complete |authentik| stack for you!
You can override any of the values of the |helm| chart as normal through this CRD.
The |helm| values must be indented under the ``values`` key as shown below.
For a full list of overrideable options please see the ak charts values.yaml file for an exhaustive list of options.

.. code-block:: yaml
   :caption: ak-sample.yaml | A simple Ak CRD with some settings you should consider

   apiVersion: akm.goauthentik.io/v1alpha1
   kind: Ak
   metadata:
     labels:
       app.kubernetes.io/name: ak
       app.kubernetes.io/instance: ak-sample
       app.kubernetes.io/part-of: operator
       app.kubernetes.io/managed-by: kustomize
       app.kubernetes.io/created-by: operator
     name: ak-sample
     namespace: auth
   spec:
     values:
       # Any child of values is taken to mean helm values for the underlying Ak helm chart
       # You may override some or all of the fields.
       # Please see chars/operator/ak/values.yaml for possible overrides
       # https://gitlab.com/GeorgeRaven/authentik-manager/-/blob/master/charts/ak/values.yaml
       # Following are some basic overrides that you should consider
       global:
         domain:
           base: example.org
           full: auth.example.org
       smtp:
         enabled: false
         username: somebody@example.org
         port: 587
         host: smtp.gmail.com
         from: noreply@example.org
       secret:
         # disabled here to allow you to load your own secret that you should definately have backed up
         generate: false
         randLength: 30
         # you probably dont want to change this name as you will have to change
         # it everywhere in subcharts
         name: auth


