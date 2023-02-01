.. include:: /substitutions

.. _section_usage:

Usage
=====

User creation
-------------

If you do not have a user yet to log in with, you can create the first admin user manually, once.
By default |authentik| has a flow to do this. To use a flow you must visit its URL.

The path portion of the URL for the default user setup is ``/if/flow/initial-setup/``, where ``initial-setup`` is the slug for the blueprint. So for instance if you set your full |authentik| domain to be ``auth.example.org`` (the default in the |helm| charts), you would want to visit ``https://auth.example.org/if/flow/initial-setup/``.

While you may have set the domain to be ``auth.example.org`` this must actually resolve to the IP of the |k8s| cluster load balancer which is actually serving |authentik|. you can check this by using the ``nslookup`` tool :code:`nslookup auth.example.org` If you are developing locally the easiest way to do this is by changing your ``/etc/hosts`` file to include something like the following:

.. code-block:: txt

   192.168.49.2 auth.example.org

.. note::

   You cannot visit the IP directly without setting the hostname. While it is possible to initiate a connection, the reverse proxy will not know which application to route you to since many can be hosted at the same IP. Thus it will just shrug you off with some error. The reverse proxy uses the domain name used in requests to then proxy you to some backend service like auth.example.org vs nextcloud.example.org might be on the same IP.

   Also note for local development you can also visit auth.example.org:30443 or any port for that matter, as long as the domain is correct. This is useful as usually a local deployment will not be on the default port 80 (http) or 443 (https). If you wanted to proxy all local requests from 443 (https) to 30443 (non standard) so that browsers play nicer, then you can use socat. :code:`socat TCP-LISTEN:443,fork TCP:192.168.49.2:30443`. This assumes the IP of the load balancer is 192.168.49.2 and is routable.

   If you are using minikube you can find this IP using :code:`minikube -n ingress-nginx service ingress-nginx-controller --url` if the ingress-nginx addon is enabled. 

If the cluster is remote / if you want others to be able to access it too, you will need to change your DNS records to point the domain name to the IP which hosts |k8s|, and from which it is available from.

Once you visit the page you will be walked through user creation via a set of text fields. Enter your desired credentials and you are ready to go.


Ingress Forward Auth
--------------------

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

