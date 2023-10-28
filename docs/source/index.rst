.. include:: substitutions

Introduction
============

Authentik-Manager (AKM) is a custom |operator| to make |authentik| and its setup declarative, expressly for the objective of making it easier to |gitops| |authentik|.

What this means as a user is that you only have to deploy the |operator| which is simple to do declaratively. The |operator| can then expose |crd|\ s to allow you to declare all your |authentik| configuration consistently. Our goal is to allow you to go as low level as individual blueprints or you can use our higher level |crd|\ s to automatically create and manage authentication proxies, forward auth, or authentication bearers. Please refer to `milestones <https://gitlab.com/GeorgeRaven/authentik-manager/-/milestones?sort=due_date_desc&state=all>`_ to see the current progress towards these goals and speculative timelines.

We use a hybrid |operator| with both Go and |helm| based controllers. This means it is as easy as using helm to configure and deploy the |authentik| specifics, but we have the low level control in Go to extend the level of automation of this |operator| beyond what |helm| could provide alone.

Key (planned) features of our work here but not exhaustive:

- Support for secrets; Many |authentik| |helm| charts use plain-text values to pass sensitive data. We support automatic generation of secrets to make it as easy to get started as possible, and later even side-loading secrets with tools like Bitnami sealed-secrets.
- Declarative configuration via blueprints; |authentik| by default does not come with a central configuration, instead you can use blueprints to add or remove functionality / behaviour. However this is quite difficult to do declaratively as usually some manual clicking is involved. We prefer it to be declarative as it is more consistent, more direct, and less error prone.
- Management of the entire |authentik| life-cycle; We love |authentik|, we want the very best for and from it, but we don't want to have to laboriously worry about it. We want backups, updates, and (re-)configuration to just happen. We know how it is, it can be hard to keep track of all the versions of all our apps, |k8s| has many many apps to keep track of. With our |operator| we aim to keep |authentik| and related resources up-to-date, eventually we want even the |operator| itself to (optionally) self-update within limits.

.. warning::

  This body of work is HIGHLY experimental. Things will break! Everything is versioned to minimise the consequences, but of course when you upgrade expect that a few little bits will need changing to conform to newer standards since everything is in alpha. I would not recommend picking this work up yet due to this uncertainty, however if you like to live life on the edge, welcome, we could use your thoughts!

Contents
========

.. toctree::
  :maxdepth: 1
  :caption: Table of Contents
  :numbered:

  license
  basics
  advanced
  operator
  glossary

Demo
====

If you want to quickly demo authentik-manager on your local desktop Linux system you will need:

- minikube
- podman
- kubectl
- helm
- make
- git
- yq
- tr

Then you can start the full cluster, AKM, and install example |crd|\ s. This will open your browser to the minikube |k8s| dashboard and provide final instructions as you go:

.. code-block::

   git clone https://gitlab.com/GeorgeRaven/authentik-manager && cd authentik-manager && make demo

Acknowledgments
===============

We would like to thank the authentik team for their amazing work, and for their tireless help!

We would also like to thank Ashley McNamara, Mat Ryer for https://gopherize.me/ which we used to generate Godrik the gopher! We would also like to thank Renee French for the original go gopher!
