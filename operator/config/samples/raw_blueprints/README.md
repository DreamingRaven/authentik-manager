# Raw Blueprints

This directory holds example authentik raw blueprints. To use these with AKM simply indent and append them to the following snippet:

```yaml

apiVersion: akm.goauthentik.io/v1alpha1
kind: AkBlueprint
metadata:
  name: <YOUR UNIQUE BLUPRINT NAME>
  namespace: <AUTHNTIK NAMESPACE>
spec:
  file: /blueprints/custom/<YOUR UNIQUE BLUPRINT NAME>.yaml
  blueprint: |
	<YOUR INDENTED BLUEPRINT HERE>
```

See more wrapped examples in operator/config/samples directory.
