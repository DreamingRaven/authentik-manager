.. include:: substitutions

Introduction
============

Authentik-Manager is a completely custom |helm| chart and |operator|. This |operator| defines |crd|\ s that are managed by |controller|\ s.
The objective of the |helm| chart is to help you deploy |authentik| to a |k8s| cluster, declaratively, primarily geared for |gitops|.

.. warning::

  This body of work is HIGHLY experimental. Things will break! Everything is versioned to minimise the consequences, but of course when you upgrade expect that a few little bits will need changing to conform to newer standards. I would not recommend picking this work up yet due to this uncertainty, however if you like to live life on the edge, welcome, we could use your thoughts!

Contents
========

.. toctree::
  :maxdepth: 1
  :caption: Table of Contents
  :numbered:

  license
  install
  usage
  operator
  upgrade
  uninstall
  glossary
