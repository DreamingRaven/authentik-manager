Authentik Helm
==============

Just add SMTP!

This is heavily a work in progress. An operator is also being developed so that configuration of Authentik can dynamically change according to the cluster. In particular when the cluster creates an ingress resource tagged with some specific label, we can tell Authentik to create a authentication proxy and potentially set the basic rules.

Installation
++++++++++++

This chart is served right here as a Gitlab helm package.

Adding our helm chart package registry on gitlab.


.. code-block:: bash

   helm repo add authentik-helm-registry https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

.. note::

   The api endpoint is: https://gitlab.com/api/v4/projects/41806964/packages/helm/api/stable/charts but the chart is actually at https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

Ensuring our local index of the helm chart is up to date.

.. code-block:: bash

   helm repo update authentik-helm-registry

Searching our package registry for available versions.

.. code-block:: bash

   helm search repo authentik-helm-registry/authentik --versions

Installing a specific version of the helm chart we would like from our search previously.

.. code-block:: bash

   helm install authentik-helm-registry/authentik --version <MAJOR.MINOR.PATCH>

#OR just install the latest out local index knows about.

.. code-block:: bash

   helm install authentik-helm-registry/authentik

Now lets install everything properly, in its own namespace and with your own values. This command does not enable SMTP as this gives you a simple proof of concept install. Once you are sure this is what you are after you will then need to replace the SMTP details with some of your own beyond this short guide. Most settings you might want to change are at the top of the values.yaml file. The big exception being images and tags.

.. code-block:: bash

   helm install authentik authentik-helm-registry/authentik --version <MAJOR.MINOR.PATCH> --create-namespace --namespace auth --set global.domain.base=<example.org> --set global.domain.full=<auth.example.org> --set global.admin.name=<somebody> --set global.admin.email=<somebody@pm.me>

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

To access pgadmin use the following commands while replacing ${CHART_NAMESPACE} with whatever namespace you have installed this chart to and ${FORWARD_PORT} to whichever port on your local machine you want it to be available from.

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

- host: ${CHART_NAMESPACE}-pgsql-hl
- port: 5432
- username: postgres
- password: $(kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.postgresPassword}" | base64 -d)

Upgrade
+++++++

Upgrade from one version to another explicitly.

.. code-block:: bash

   helm upgrade authentik authentik-helm-registry/authentik --namespace auth --version <MAJOR.MINOR.PATCH>

Uninstall
+++++++++

Uninstall the helm chart and its resources but not anything that you have installed on top.

.. code-block:: bash

   helm uninstall authentik --namespace auth

Authentik Operator
++++++++++++++++++

The Authentik operator is a custom operator which currently consists of a controller for AkServer (WIP), AkWorker (WIP), AkProvider (WIP), AkOutpost (WIP), AkApplication (WIP) resources. 

We have tried to keep as much of the terminology to match that which existing Authentik users would understand.

The current proposals for CRDs are:

AkServer
--------

.. code-block:: yaml

   apiVersion:  v1alpha1
   kind:        AkServer
   metadata:
      name:     AkServer
   spec:

AkWorker
--------

.. code-block:: yaml

   apiVersion:  v1alpha1
   kind:        AkWorker
   metadata:
      name:     AkWorker
   spec:

AkProvider
----------

.. code-block:: yaml

   apiVersion:  v1alpha1
   kind:        AkProvider
   metadata:
      name:     Provider
   spec:
      consentFlow:      default-provider-authorization-explicit-consent
      # AppForwardAuth, or DomainForwardAuth
      type:     AppForwardAuth
      url:      https://app.example.com

AkOutpost
---------

.. code-block:: yaml

   apiVersion:  v1alpha1
   kind:        AkOutpost
   metadata:
      name:     Outpost
   spec:

AkApplication
---------

.. code-block:: yaml

   apiVersion:  v1alpha1
   kind:        AkApplication
   metadata:
      name:     AkApp
   spec:
      # internal application name for urls
      slug:     myapp
      # (Optional) group and show applications together in UI
      group:    nil
      # provider to handle this application
      provider: AkProvider
      # either any or all policies must match to grant access
      policyEngineMode: any
      # (Optional) UI settings for this application
      ui:
         # optional specifier for url to launch, will default to providers url if empty
         launchURL:     https://app.example.com
         icon:          https://cdn.example.com/appIcon.png
         publisher:     Organisation
         description:   "Some app of ours that does the thing."
