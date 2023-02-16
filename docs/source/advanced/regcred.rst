.. include:: /substitutions

.. _section_regcred:

Private Registry Credentials
============================

.. note::

   You won't need this for any of the works in this repository. We added this as a helpful reminder for ourselves and others that may wish to GitOps integrate authentik-manager with other works that arent openly available.

|k8s| can pull from a private registry by pointing a pod to a secret. We can tell a pod which secrets to look for using the ``pod.spec.imagePullSecrets[]`` array. Which in YAML looks something like the following highlighted lines in a minimal pod definition:

.. code-block:: yaml
   :caption: Pod definition
   :emphasize-lines: 8,9,10

   apiVersion: v1
   kind: Pod
   metadata:
     name: pod-with-private-containers
   spec:
     containers:
     - name: private-container
       image: registry.example.org
     imagePullSecrets:
     - name: regcred

So now we know how to point a pod to a secret which should contain registry credentials for (in our example) our private registry ``registry.example.org``. This secret in the example we use here is called ``regcred``. ``regcred`` must be in the same namespace as the pod (if empty defaults to ``default``), otherwise it will not be able to find and mount it. So now what does this ``regcred`` contain? Well its almost like any other secret except it follows a slightly more strict format. Lets look at a specific example:

.. code-block:: yaml
   :caption: Registry credentials secret regcred
   :emphasize-lines: 4,6

   apiVersion: v1
   kind: Secret
   metadata:
     name: regcred
   data:
     .dockerconfigjson: ewoJImF1dGhzIjogewoJCSJyZWdpc3RyeS5leGFtcGxlLmNvbSI6IHsKCQkJImF1dGgiOiAiZFhObGNqcHdZWE56ZDI5eVpBbz0iCgkJfQoJfQp9Cg==

So this is what our ``regcred`` secret looks like! The three important things to note are:

- The name matches the name in our pod definition.
  The key which holds the authentication json is ``.dockerconfigjson``
- The value of this key is a base64 encoded value which contains the entirety of our authentication json.

The reason the value is base64 encoded is because yaml is sensitive to newlines and special characters so we base64 encoded it so it does not mess with the YAML format.
If we take a look at what this base64 encoded value is, it is the following:

.. code-block:: json
   :caption: auth.json
   :emphasize-lines: 4

   {
           "auths": {
                   "registry.example.org": {
                           "auth": "dXNlcjpwYXNzd29yZAo="
                   }
           }
   }

You will note this ``auth.json`` file is json that maps a registry with some credentials. You will also note that the credentials are again base64 encoded (the give away is the ``=`` at the end).

Ok one last time what is in that credential which is encoded in base64:

.. code-block:: json
   :caption: login credential

   user:password

So really this entire chain of base64 encoding, json, secret, and imagePullSecrets is all to represent this one tiny credential and the registry that it unlocks. Knowing all this you can create this programmatically, like inside helm for instance!

:See Also: https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/

Shortcuts
---------

While the above is useful for a whole understanding of registry credentials, one can shortcut their way to a full and properly formatted ``regcred`` using the following:

.. code-block:: bash
   :caption: Registry credential generating shortcut

   podman login registry.example.org --authfile ${HOME}/.docker/config.json
   kubectl create secret -n default generic regcred --from-file .dockerconfigjson=${HOME}/.docker/config.json

:substitutes:

  - ``registry.example.org``: Your private registry e.g registry.gitlab.com
  - ``${HOME}/.docker/config.json``: Arbitrary path where podman will store its login info
  - ``default``: Namespace the credential will be generated for

