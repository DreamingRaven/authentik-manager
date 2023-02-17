.. include:: /substitutions

.. _section_install:

Install
=======

We assume you want to get started with AKM, however if you want to install just |authentik| then we also have subsequent instructions for that. Please choose one from either :ref:`section_install_akm` or :ref:`section_install_ak`, it is not necessary and not recommended to install both.

.. _section_install_akm:

Authentik Manager Install
-------------------------

AKM is installed via helm chart.
You will need:

- A functioning |k8s| cluster
- Permissions to create AKM inside the cluster
- Kubectl to communicate with the cluster
- |helm| to actually install the |helm| chart

To add the helm repo:

.. code-block:: bash

   helm repo add akm-registry https://gitlab.com/api/v4/projects/41806964/packages/helm/stable
   helm repo update akm-registry

Choose the version of AKM you want to install:

.. code-block:: bash

   helm search repo akm-registry/akm --versions

Then you can install your favoured version:

.. code-block:: bash

   helm install akm akm-registry/akm --version MAJOR.MINOR.PATCH

Congratulations that's it! Go straight to :ref:`section_usage`

.. _section_install_ak:

Authentik Install
-----------------

This chart is served right here as a Gitlab helm package.

Adding our helm chart package registry on gitlab.


.. code-block:: bash

   helm repo add akm-registry https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

.. note::

   The api endpoint is: https://gitlab.com/api/v4/projects/41806964/packages/helm/api/stable/charts but the chart is actually at https://gitlab.com/api/v4/projects/41806964/packages/helm/stable

Ensuring our local index of the helm chart is up to date.

.. code-block:: bash

   helm repo update akm-registry

Searching our package registry for available versions.

.. code-block:: bash

   helm search repo akm-registry/ak --versions

Installing a specific version of the helm chart we would like from our search previously.

.. code-block:: bash

   helm install ak akm-registry/ak --version MAJOR.MINOR.PATCH

#OR just install the latest our local index knows about.

.. code-block:: bash

   helm install ak akm-registry/ak

Now lets install everything properly, in its own namespace and with your own values. This command does not enable SMTP as this gives you a simple proof of concept install. Once you are sure this is what you are after you will then need to replace the SMTP details with some of your own beyond this short guide. Most settings you might want to change are at the top of the values.yaml file. The big exception being images and tags.

.. code-block:: bash

   helm install authentik akm-registry/ak --version MAJOR.MINOR.PATCH --create-namespace --namespace auth --set global.domain.base=example.org --set global.domain.full=auth.example.org --set global.admin.name=somebody --set global.admin.email=somebody@pm.me


.. note::

   The flags for global.admin.name and global.admin.email do not currently propagate through so you wont be able to log in with these credentials. Instead please initialise your user manually for now. See |section_usage| for how, there is inbuilt functionality for this by authentik so its just a matter of visiting a URL and giving it the details once.

.. warning::

   It will take some time for authentik to become ready, in particular it is usually Redis that takes the longest initial setup time. So do not be surprised if it is crash looping because Redis host is not found or unreachable.

By default this proof-of-concept deployment will create randomised passwords and secrets. If you want to take this from PoC to production consider using bitnami sealed-secrets, while disabling secret generation in this chart. That way nothing will start until bitnami creates the secret in the same namespace as authentik and you can save (while encrypted) the sealed-secret while keeping it git versioned. Please also note one should enable SMTP so that authentik can be completely stateless, and so users can reset their own passwords.
