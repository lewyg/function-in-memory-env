# Example manifests

You can run your function locally and test it using `crossplane beta render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

```shell
# Then, in another terminal, call it with these example manifests
$ crossplane beta render xr.yaml composition.yaml functions.yaml --context-values=apiextensions.crossplane.io/environment='{"key": "value"}' -r
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
---
apiVersion: apiextensions.crossplane.io/v1alpha1
data:
  key: value
kind: EnvironmentConfig
metadata:
  annotations:
    crossplane.io/composition-resource-name: in-memory-env
  labels:
    crossplane.io/composite: example-xr
    xr-apiversion: example.crossplane.io_v1
    xr-kind: XR
    xr-name: example-xr
```
