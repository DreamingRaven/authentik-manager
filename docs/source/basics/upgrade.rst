.. include:: /substitutions

.. _section_upgrade:

Upgrade
=======

Depending on your choice of installation you should use one set of the following upgrade instructions either :ref:`section_upgrade_akm`, or :ref:`section_upgrade_ak`.

.. _section_upgrade_akm:

Authentik-Manager Upgrade
-------------------------

Upgrade from one version of the |operator| to another explicitly, however this is only the operators version. The operator will manage |authentik| and keep it up to date where possible.

.. code-block:: bash

   helm upgrade akm akm-registry/akm --namespace auth --version MAJOR.MINOR.PATCH

.. _section_upgrade_ak:

Authentik Upgrade
-----------------

Upgrade from one version of |authentik| to another explicitly this will be static until you manually bump the version.

.. code-block:: bash

   helm upgrade ak akm-registry/ak --namespace auth --version MAJOR.MINOR.PATCH

