Authentik Operator
==================

The authentik-operator serves to manage an existing authentik server and worker. In particular this operator automates deployment and management of applications, providers, and outposts, along with the necessary ingress annotations to link existing services to the authentik server via the aforementioned application, providers and outposts.

Prerequisites
-------------

You will need a configured and running Kubernetes cluster. In most instances while testing, this will be a minikube instance, but it can be any cluster. The key requirement is that it is configured in kubeconfig file. Further the default / current context is what will be used. You can use `kubectl cluster-info` to show what the default is.

Currently you will also need a running authentik server and worker to control with this operator. As this operator becomes more advanced it may also include the ability to create these for you, but for now treat it as you would a developer who links your applications using CRDs you create, with the abstractions internal to authentik without you needing to click through a plethora of menus. This way it is an operator in the original sense.

The hope is that this will make it much easier to consistently and repeatably configure and set-up clusters in a GitOps fashion, by making authentik configuration kube native.

What is an Operator
-------------------

This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/) 
which provides a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster 

Manually Running the Operator
-----------------------------
0. Generate the api resources and manifests

.. code-block::

   make generate manifests

1. Install the CRDs into the cluster:

.. code-block::

   make install

2. Run your controller in the foreground:

.. code-block::

   make run
