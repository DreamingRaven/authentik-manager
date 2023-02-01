.. include:: /substitutions

.. _section_uninstall:

Uninstall
=========

Uninstall the helm chart and its resources but not anything that you have installed on top.

.. code-block:: bash

   helm uninstall authentik --namespace auth

.. note::

   We are still working on ensuring everything is cleaned up. In most cases you should be fine, but I would double check both the authentik-worker deployment and any affected ingress resources for any lingering volumes or forward-auths that you may not want any more.

