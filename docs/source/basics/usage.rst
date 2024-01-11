.. include:: /substitutions

.. _section_usage:

Usage
=====

Now depending on your chosen installation, whether |operator| or raw |authentik| helm chart, please choose one of the following subsections :ref:`section_usage_akm`, or :ref:`section_usage_ak`.

.. _section_usage_akm:

Authentik-Manager Usage
-----------------------

By default AKM does nothing. You will just see the default AKM pods, as they do not have any instructions with anything to do yet.

There are three things you will need to do.

- Generate a secret that will be used by the |authentik| stack. AKM will generate the stack for you but it cannot generate a default secret for you. Please make your own and keep it safe!
- Tell AKM what sort of |authentik| stack you would like using the Ak CRD.
- Tell AKM what you want to integrate into the |authentik| stack using higher level CRDs (WIP).

Declare a Secret
++++++++++++++++

.. note::

   Ideally we would have generated this secret for you, but a limitation of the operator SDK with hybrid plugins are that helm controllers cannot use the lookup function for us to generate and persist passwords for you.

Authentik, redis, and postgres all use passwords between each other. We need to define a secret with all the passwords they will need. It must be called ``auth`` be in whatever namespace AKM is in, and must have the keys as per the following example. Feel free to use this one that I generated for testing, but please change all these values in production to something different.

.. code-block:: yaml
   :caption: auth.yaml | auth secret with base64 encoded passwords

   apiVersion: v1
   kind: Secret
   metadata:
     name: auth # must be named auth for now
     namespace: auth # ensure this matches akms namespace usually also auth
   type: Opaque
   data:
     authDuoApiKey: UUVXUWhUSTA2MXVaWWVLMlhCbEkzVE5IZDdzcWlC
     authJwtToken: WG5taGl6aFRvcWdqYlhkWXlxY2s4QXgzeGFBTUFh
     authSessionEncryptionKey: MXNqZ0pwVWdkdXNKSXdnT3dqb0FtV3JkOVVzQ0hC
     authStorageEncryptionKey: NmZ2a1VQVVZsVVFMY2NYVjRudU1nRjEyT21CaTl2
     ldapAdminPassword: cUpLZ1RtQVNKaEhiRUhGT1NVY3FGM2NsaGhOcE9C
     oidcHmacSecret: eG1sN2k0TEYxTk1sN3ZDOHhEYWh6U1VYaDZLTG1J
     oidcPrivateKey: Mm5NVFNNYW4zSUEwcGU4UXd6Y3huQ3FZdGNFVUlw
     pgAdminPassword: MVRaNm9iWEI4Sk50V0NHR1FCd3NNTjM1WjdPQ210
     postgresPassword: TUl3SHNja1NxaENsaTBLQ0VtcTVSWkRsZDc0NHZQ
     postgresReplicationPassword: cGczYXVSd2VhcnNSc3QwSVBNQmpCVUhMMFVaYVBz
     postgresUserPassword: WHB4aFJXY0l3NnhiNUFLbzhuWjdCWnJvR0dpZTl3
     redisPassword: dnZyNmtJTFkweEEwMVlidlZJajN6OXY2SmZPSENk
     redisSentinelPassword: dVR3bTdZdE4ydjh5cm5EY2RteWNMb1lyZjNhaVRp
     sessionSecret: cFN4U1lZQVlncG1JVmlPek9hTVJkdjJZaTVkQ21t
     smtpPassword: WWl4N1BXQVZLQ1JvSjdaRzF6U2QxT3FBVWlGV1F6
     smtpUsername: WWl4N1BXQVZLQ1JvSjdaRzF6U2QxT3FBVWlGV1F6

.. code-block:: bash
   :caption: example generating a single base64 encoded secret

   echo "someReallyLongRandomPassword" | base64

Once you have auth.yaml from the above example and your own passwords you can install it using kubectl:

.. code-block:: bash
   :caption: installing auth.yaml

   kubectl apply -f auth.yaml

You can manage the secret manually but there are lots of tools already available to manage secrets for you. Consider using Bitnami sealed secrets which decrypts and side-loads secrets for you, and is also |gitops| compatible.

Declare an Ak CRD
+++++++++++++++++

Now that the secret exists we can tell AKM to create an |authentik| instance for us using the Ak CRD. Some examples of the Ak CRD can be found in ``operator/config/samples/*_ak.yaml``.

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


Then apply it with:

.. code-block:: bash

   kubectl apply -f ak-sample.yaml

This will create a complete |authentik| stack for you!
You can override any of the values of the |helm| chart as normal through this CRD.

.. note::

   While you can enable secret generation it is highly discouraged, as the operator plugin we use to hybridise the sdk consumes an inordinate amount of resources trying to reconcile what is already reconciled.

Authentik should be created for you, simply visit ``https://<global.domain.full>/if/flow/initial-setup/``. If you have not got DNS set up you may need to connect directly by editing your ``/etc/hosts`` file to add the line ``<minikube ip> <global.domain.full>``. Remember to replace minikube ip with minikubes actual ip address which you can find with ``minikube ip`` command. Also replace ``global.domain.full`` to whatever you set it as in the Ak crd you just applied.


Declare an Application CRD
++++++++++++++++++++++++++

Your authentik installation should be ready to go. There are a few example CRs in ``operator/config/samples/akm_*`` which can do various blueprint things like changing the logos, setting up a default tenant, etc. We are still working on OIDC abstractions but the akblueprint should work as expected which you can use to declare almost anything for the time being while we set up more automation for things like OIDC.

.. _section_usage_ak:

Authentik Usage
---------------

If you installed |authentik| directly via the static helm chart you will need to know the following:

User creation
+++++++++++++

If you do not have a user yet to log in with, you can create the first admin user manually, once.
By default |authentik| has a flow to do this. To use a flow you must visit its URL.

The path portion of the URL for the default user setup is ``/if/flow/initial-setup/``, where ``initial-setup`` is the slug for the blueprint. So for instance if you set your full |authentik| domain to be ``auth.example.org`` (the default in the |helm| charts), you would want to visit ``https://auth.example.org/if/flow/initial-setup/``.

While you may have set the domain to be ``auth.example.org`` this must actually resolve to the IP of the |k8s| cluster load balancer which is actually serving |authentik|. you can check this by using the ``nslookup`` tool :code:`nslookup auth.example.org` If you are developing locally the easiest way to do this is by changing your ``/etc/hosts`` file to include something like the following:

.. code-block::

   192.168.49.2 auth.example.org

.. note::

   You cannot visit the IP directly without setting the hostname. While it is possible to initiate a connection, the reverse proxy will not know which application to route you to since many can be hosted at the same IP. Thus it will just shrug you off with some error. The reverse proxy uses the domain name used in requests to then proxy you to some backend service like auth.example.org vs nextcloud.example.org might be on the same IP.

   Also note for local development you can also visit auth.example.org:30443 or any port for that matter, as long as the domain is correct. This is useful as usually a local deployment will not be on the default port 80 (http) or 443 (https). If you wanted to proxy all local requests from 443 (https) to 30443 (non standard) so that browsers play nicer, then you can use socat. :code:`socat TCP-LISTEN:443,fork TCP:192.168.49.2:30443`. This assumes the IP of the load balancer is 192.168.49.2 and is routable.

   If you are using minikube you can find this IP using :code:`minikube -n ingress-nginx service ingress-nginx-controller --url` if the ingress-nginx addon is enabled.

If the cluster is remote / if you want others to be able to access it too, you will need to change your DNS records to point the domain name to the IP which hosts |k8s|, and from which it is available from.

Once you visit the page you will be walked through user creation via a set of text fields. Enter your desired credentials and you are ready to go.


Ingress Forward Auth
++++++++++++++++++++

Now that the chart is installed it will do... Nothing. Unless your ingress resource is configured to rely on |authentik| for authentication, everything in your cluster will not be affected.
One of the more common ways to do this is via the forward auth header.

Each ingress resource is configured individually to listen to authentik but not authentiks own ingress resource. Please add the following annotations to your ingress-nginx ingress resource to have it listen to authentik. Note that you must point the auth-url to an outpost that knows of the ingress host.

.. note::

   We are building a |crd| to do this for you. Otherwise this can be very tedious and error prone, especially when combined with having to click through a series of menus to let authentik know about this forward auth happening.

.. code-block:: yaml

   #Additional annotations necessary to have authentik be an authentication middleware on the nginx proxy.

   annotations:
      nginx.ingress.kubernetes.io/auth-url: http://{ OUTPOST SERVICE}.{ OUTPOST NAMESPACE }.svc.cluster.local:9000/outpost.goauthentik.io/auth/nginx
      nginx.ingress.kubernetes.io/auth-signin: https://{ INGRESS HOST OF YOUR APP }/outpost.goauthentik.io/start?rd=$escaped_request_uri
      nginx.ingress.kubernetes.io/auth-response-headers: Set-Cookie,X-authentik-username,X-authentik-groups,X-authentik-email,X-authentik-name,X-authentik-uid
      nginx.ingress.kubernetes.io/auth-snippet: proxy_set_header X-Forwarded-Host $http_host;

PGAdmin
+++++++

To access pgadmin use the following commands while replacing CHART_NAMESPACE with whatever namespace you have installed this chart to and FORWARD_PORT to whichever port on your local machine you want it to be available from.

.. code-block:: bash

   # wait for the pgadmin deployment to come alive
   kubectl wait --timeout=600s --for=condition=Available=True -n ${CHART_NAMESPACE} deployment pgadmin
   # get username / email to log in with
   kubectl -n ${CHART_NAMESPACE} get deployment pgadmin-deployment -o jsonpath="{.spec.template.spec.containers[0].env[0].value}"
   # get the user password
   kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.pgAdminPassword}" | base64 -d && echo
   # expose pgadmin locked inside the cluster to a port of our choice e.g localhost:8079
   kubectl port-forward svc/pgadmin-service -n ${CHART_NAMESPACE} ${FORWARD_PORT}:http-port

Once logged in you can add the postgres service running in the cluster:

- host: postgres
- port: 5432
- db: postgres
- username: postgres
- password: ``$(kubectl -n ${CHART_NAMESPACE} get secret auth -o jsonpath="{.data.postgresPassword}" | base64 -d)``

