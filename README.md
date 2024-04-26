# function-in-memory-env
[![CI](https://github.com/lewyg/function-in-memory-env/actions/workflows/ci.yml/badge.svg)](https://github.com/lewyg/function-in-memory-env/actions/workflows/ci.yml)

Function automatically creates [EnvironmentConfig](https://docs.crossplane.io/latest/concepts/environment-configs/) that represents in-memory environment of XR.

Add it as a last step in pipeline, to include all patches from previous steps:
```yaml
- step: in-memory-config
  functionRef:
    name: function-in-memory-env
```

then to create in-memory env config, set this annotation of XR: `inmemoryenv.fn.crossplane.io/enabled: "true"`
(this function is disabled by default, will only create EnvConfig when annotation is set)

### Example:

Given environment configs:
```yaml
---
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: EnvironmentConfig
metadata:
  labels:
    test: test
  name: test-env-config
data:
  key1: value1
  key2:
    key2nested: value2
---
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: EnvironmentConfig
metadata:
  labels:
    test: test
  name: test-env-config-2
data:
  key1: value1overwritten
---
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: EnvironmentConfig
metadata:
  labels:
    test: test
  name: test-env-config-3
data:
  key3: value3
  key2:
    key2nested2: value2updated
```

and selector on XR:
```yaml
- type: Selector
  selector:
    mode: Multiple
    matchLabels:
      - key: test
        type: Value
        value: test
```

the function creates new EnvConfig for the XR:
```yaml
apiVersion: apiextensions.crossplane.io/v1alpha1
kind: EnvironmentConfig
data:
  key1: value1overwritten
  key2:
    key2nested: value2
    key2nested2: value2updated
  key3: value3
...
```


## Developing this function

* [Follow the guide to writing a composition function in Go][function guide]
* [Learn about how composition functions work][functions]
* [Read the function-sdk-go package documentation][package docs]

This template uses [Go][go], [Docker][docker], and the [Crossplane CLI][cli] to
build functions.

```shell
# Run code generation - see input/generate.go
$ go generate ./...

# Run tests - see fn_test.go
$ go test ./...

# Build the function's runtime image - see Dockerfile
$ docker build . --tag=runtime

# Build a function package - see package/crossplane.yaml
$ crossplane xpkg build -f package --embed-runtime-image=runtime
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli
