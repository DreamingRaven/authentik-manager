.. include:: /substitutions

Install
=======

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

   helm install authentik-manager-registry/authentik --version MAJOR.MINOR.PATCH

#OR just install the latest out local index knows about.

.. code-block:: bash

   helm install authentik-manager-registry/authentik

Now lets install everything properly, in its own namespace and with your own values. This command does not enable SMTP as this gives you a simple proof of concept install. Once you are sure this is what you are after you will then need to replace the SMTP details with some of your own beyond this short guide. Most settings you might want to change are at the top of the values.yaml file. The big exception being images and tags.

.. code-block:: bash

   helm install authentik authentik-manager-registry/authentik --version MAJOR.MINOR.PATCH --create-namespace --namespace auth --set global.domain.base=example.org --set global.domain.full=auth.example.org --set global.admin.name=somebody --set global.admin.email=somebody@pm.me

.. warning::

   It will take some time for authentik to become ready, in particular it is usually Redis that takes the longest initial setup time. So do not be surprised if it is crash looping because Redis host is not found or unreachable.

By default this proof-of-concept deployment will create randomised passwords and secrets. If you want to take this from PoC to production consider using bitnami sealed-secrets, while disabling secret generation in this chart. That way nothing will start until bitnami creates the secret in the same namespace as authentik and you can save (while encrypted) the sealed-secret while keeping it git versioned. Please also note one should enable SMTP so that authentik can be completely stateless, and so users can reset their own passwords.
