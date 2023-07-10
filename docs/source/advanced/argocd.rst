.. include:: /substitutions

.. _section_argocd:

ArgoCD
======

ArgoCD is as per their own description a declarative, |gitops| continuous delivery tool.

ArgoCD is the primary targeted |gitops| tool, although we add generic implementations and labels to enable as wide a variety of support as possible.

Resource Tacking
----------------

It is important to note that ArgoCD is more than capable out of the box of tracking AKM itself, but it needs some assistance tracking the operator + helm deployed AK resources that are subsequently deployed by AKM.

You will notice that when you deploy applications with ArgoCD, they have additional labels (and possibly annotations) to those strictly defined in your manifests. When you deploy AKM by default as an ArgoCD application you will see an additional label ``argocd.argoproj.io/instance: <YOUR ARGOCD APPLICATION NAME>``. ArgoCD also supports the more generic recommended |k8s| label ``app.kubernetes.io/instance: <YOUR ARGOCD APPLICATION NAME>`` which we also must propagate to ensure a wider variety of support.

To allow ArgoCD to know that the deployed AK resources are related to the ArgoCD application of AKM we need to use these labels. The operator will automatically populate the necessary labels by checking itself for these labels and propagating them into the deployed AK helm charts and dependency charts.

See Also
--------

- ArgoCD resource tracking https://argo-cd.readthedocs.io/en/stable/user-guide/resource_tracking/
- Recommended labels https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/
- Labels and Selectors https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/
