.. include:: /substitutions

.. _section_uninstall:

Uninstall
=========

To uninstall either the |operator| or |authentik| depending on which version you have installed choose either :ref:`section_uninstall_akm`, or :ref:`section_uninstall_ak`.

.. _section_uninstall_akm:

Authentik-Manager Uninstall
---------------------------

Uninstall the |operator|, this will cause |k8s| to delete all resources dependent on this operator.

.. code-block:: bash

   helm uninstall akm --namespace auth

.. _section_uninstall_ak:

Authentik Uninstall
-------------------

Uninstall the helm chart and its resources but not anything that you have installed on top.

.. code-block:: bash

   helm uninstall ak --namespace auth

.. note::

   We are still working on ensuring everything is cleaned up. In most cases you should be fine, but I would double check both the authentik-worker deployment and any affected ingress resources for any lingering volumes or forward-auths that you may not want any more.

