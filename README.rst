Authentik Manager
=================

Just add SMTP!

Authentik-Manager is a custom authentik-helm chart with an additional operator. The helm chart helps deploy the base system, and the operator with its additional CRDs makes it easier to declaratively define your authentik setup.

Please note this is still heavily a work in progress, that has only recently started. If you like living life on the edge welcome, otherwise good news on progress should come soon.

Installation
++++++++++++

This chart is served right here as a Gitlab helm package.

Adding our helm chart package registry on gitlab.


.. code-block:: bash

   helm repo add authentik-manager-registry https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

.. note::

   The api endpoint is: https://gitlab.com/api/v4/projects/41806964/packages/helm/api/stable/charts but the chart is actually at https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

Ensuring our local index of the helm chart is up to date.

.. code-block:: bash

   helm repo update authentik-manager-registry

Searching our package registry for available versions.

.. code-block:: bash

   helm search repo authentik-manager-registry/authentik --versions

Installing a specific version of the helm chart we would like from our search previously.

.. code-block:: bash

   helm install authentik-manager-registry/authentik --version <MAJOR.MINOR.PATCH>

#OR just install the latest out local index knows about.

.. code-block:: bash

   helm install authentik-manager-registry/authentik

Now lets install everything properly, in its own namespace and with your own values. This command does not enable SMTP as this gives you a simple proof of concept install. Once you are sure this is what you are after you will then need to replace the SMTP details with some of your own beyond this short guide. Most settings you might want to change are at the top of the values.yaml file. The big exception being images and tags.

.. code-block:: bash

   helm install authentik authentik-manager-registry/authentik --version <MAJOR.MINOR.PATCH> --create-namespace --namespace auth --set global.domain.base=<example.org> --set global.domain.full=<auth.example.org> --set global.admin.name=<somebody> --set global.admin.email=<somebody@pm.me>

.. warning::

   It will take some time for authentik to become ready, in particular it is usually Redis that takes the longest initial setup time. So do not be surprised if it is crash looping because Redis host is not found or unreachable.

By default this proof-of-concept deployment will create randomised passwords and secrets. If you want to take this from PoC to production consider using bitnami sealed-secrets, while disabling secret generation in this chart. That way nothing will start until bitnami creates the secret in the same namespace as authentik and you can save (while encrypted) the sealed-secret while keeping it git versioned. Please also note one should enable SMTP so that authentik can be completely stateless, and so users can reset their own passwords.

Usage
+++++

Now that the chart is installed it will do... Nothing. Unless your ingress controller is configured to rely on authentik for authentication everything in your cluster will not be affected.

Each ingress resource is configured individually to listen to authentik but not authentiks own ingress resource. Please add the following annotations to your ingress-nginx ingress resource to have it listen to authentik. Note that you must point the auth-url to an outpost that knows of the ingress host.

.. code-block:: bash

   #Additional annotations necessary to have authentik be an authentication middleware on the nginx proxy.

   annotations:
      nginx.ingress.kubernetes.io/auth-url: http://{{ OUTPOST SERVICE}}.{{ OUTPOST NAMESPACE}}.svc.cluster.local:9000/outpost.goauthentik.io/auth/nginx
      nginx.ingress.kubernetes.io/auth-signin: https://{{ INGRESS HOST OF YOUR APP }}/outpost.goauthentik.io/start?rd=$escaped_request_uri
      nginx.ingress.kubernetes.io/auth-response-headers: Set-Cookie,X-authentik-username,X-authentik-groups,X-authentik-email,X-authentik-name,X-authentik-uid
      nginx.ingress.kubernetes.io/auth-snippet: proxy_set_header X-Forwarded-Host $http_host;

TODO: note that an ingress resource must exist that points to authentik for every SSOed app which has a path for {{ APP DOMAIN }}/outpost.goauthentik.io/

PGAdmin
-------

To access pgadmin use the following commands while replacing CHART_NAMESPACE with whatever namespace you have installed this chart to and FORWARD_PORT to whichever port on your local machine you want it to be available from.

.. code-block:: bash

   # wait for the pgadmin deployment to come alive
   kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pgadmin-deployment
   # get username / email to log in with
   kubectl -n ${CHART_NAMESPACE} get deployment pgadmin-deployment -o jsonpath="{.spec.template.spec.containers[0].env[0].value}"
   # get the user password
   kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
   # expose pgadmin locked inside the cluster to a port of our choice e.g localhost:8079
   kubectl port-forward svc/pgadmin-service -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http-port

Once logged in you can add the postgres service running in the cluster:

- host: ``${CHART_NAMESPACE}-pgsql-hl``
- port: 5432
- username: postgres
- password: ``$(kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.postgresPassword}" | base64 -d)``

Upgrade
+++++++

Upgrade from one version to another explicitly.

.. code-block:: bash

   helm upgrade authentik authentik-manager-registry/authentik --namespace auth --version <MAJOR.MINOR.PATCH>

Uninstall
+++++++++

Uninstall the helm chart and its resources but not anything that you have installed on top.

.. code-block:: bash

   helm uninstall authentik --namespace auth

Authentik Operator
++++++++++++++++++

The authentik `operator <https://kubernetes.io/docs/concepts/extend-kubernetes/operator/>`_ is a simple `controller <https://kubernetes.io/docs/concepts/architecture/controller/>`_ for some of our own `CRDs <https://kubernetes.io/docs/tasks/extend-kubernetes/custom-resources/custom-resource-definitions/>`_. At its core this operator simply seeks to make it possible to declaratively configure authentik. The standard way to configure authentik beyond the environment variables they already expose are `blueprints <https://goauthentik.io/developer-docs/blueprints/>`_. Blueprints are how most components in authentik can be customised and adapted like two-factor authentication support, and changing backgrounds etc. The issue is, it is difficult to maintain such blueprints in a kubernetes native manner with tools such as GitOps, since much of the configuration of authentik is otherwise done with ClikOps. It should be noted that it is still possible, it is just significantly easier to defined CRDs that can exist in any helm-chart and have the operator automatically reconcile it into and connect with authentik. A good example is a standalone app that needs authentication, it can have its own specification for proxy, forward-auth, or authentication bearers in with the rest of the definition for the app. Then when that helm chart is installed it will be reconciled according to the CRDs included in with that helm chart, rather than then having to go through authentik and ClickOps your way to connecting your new app to authentik.

We have tried to keep as much of the terminology to match that which existing Authentik users would understand, with blueprints, providers, outposts etc.

The operator is deployed as just another container in your cluster with service accounts, roles, clusterRoles, roleBindings and ClusterRoleBindings limited to the minimum requirements of the reconciliation loop so that it cannot access anything other than its own CRDs and ingress resources. To make things easier the helm-chart CRDs are templated in ``charts/crds``, the operator helm-chart is in ``charts/operator``, and a meta chart that also installs authentik is in ``charts/auth``.

To see all current CRDs the current specification can be found in the ``operator/api/v1alpha1`` directory. Example CRDs can be found in ``operator/config/samples``. The specific controllers for these resources can be found in ``operator/controllers``

Please also be aware only AkBlueprint is semi-complete. AkBlueprint will back almost all other CRDs as a low level interface, while the other CRDs will deploy much higher level concepts that require blueprints to exist.

For more operator specific details please see the ``operator`` directory README.

Documentation
+++++++++++++

WIP!

https://georgeraven.gitlab.io/authentik-manager/
