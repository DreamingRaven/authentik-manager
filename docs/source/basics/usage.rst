.. include:: /substitutions

Usage
=====

Now that the chart is installed it will do... Nothing. Unless your ingress controller is configured to rely on authentik for authentication everything in your cluster will not be affected.

Each ingress resource is configured individually to listen to authentik but not authentiks own ingress resource. Please add the following annotations to your ingress-nginx ingress resource to have it listen to authentik. Note that you must point the auth-url to an outpost that knows of the ingress host.

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

