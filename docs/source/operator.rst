Authentik Operator
==================

The authentik `operator <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>`_ is a simple `controller <https://kubernetes.io/docs/concepts/architecture/controller/>`_ for some of our own `CRDs <https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/>`_. At its core this operator simply seeks to make it possible to declaratively configure authentik. The standard way to configure authentik beyond the environment variables they already expose are `blueprints <https://goauthentik.io/developer-docs/blueprints/>`_. Blueprints are how most components in authentik can be customised and adapted like two-factor authentication support, and changing backgrounds etc. The issue is, it is difficult to maintain such blueprints in a kubernetes native manner with tools such as GitOps, since much of the configuration of authentik is otherwise done with ClikOps. It should be noted that it is still possible, it is just significantly easier to defined CRDs that can exist in any helm-chart and have the operator automatically reconcile it into and connect with authentik. A good example is a standalone app that needs authentication, it can have its own specification for proxy, forward-auth, or authentication bearers in with the rest of the definition for the app. Then when that helm chart is installed it will be reconciled according to the CRDs included in with that helm chart, rather than then having to go through authentik and ClickOps your way to connecting your new app to authentik.

We have tried to keep as much of the terminology to match that which existing Authentik users would understand, with blueprints, providers, outposts etc.

The operator is deployed as just another container in your cluster with service accounts, roles, clusterRoles, roleBindings and ClusterRoleBindings limited to the minimum requirements of the reconciliation loop so that it cannot access anything other than its own CRDs and ingress resources. To make things easier the helm-chart CRDs are templated in ``charts/crds``, the operator helm-chart is in ``charts/operator``, and a meta chart that also installs authentik is in ``charts/auth``.

To see all current CRDs the current specification can be found in the ``operator/api/v1alpha1`` directory. Example CRDs can be found in ``operator/config/samples``. The specific controllers for these resources can be found in ``operator/controllers``

Please also be aware only AkBlueprint is semi-complete. AkBlueprint will back almost all other CRDs as a low level interface, while the other CRDs will deploy much higher level concepts that require blueprints to exist.

For more operator specific details please see the ``operator`` directory README.

