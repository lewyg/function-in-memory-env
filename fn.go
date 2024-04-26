package main

import (
	"context"
	"slices"
	"strings"

	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/json"

	xpv1alpha1 "github.com/crossplane/crossplane/apis/apiextensions/v1alpha1"
	"github.com/crossplane/function-sdk-go/errors"
	"github.com/crossplane/function-sdk-go/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

const (
	functionContextKeyEnvironment = "apiextensions.crossplane.io/environment"

	annotationKeyInMemoryEnvEnabled = "inmemoryenv.fn.crossplane.io/enabled"
)

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	// Get the desired composite resource from the request.
	observedComposite, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	if inMemoryEnvEnabled, found := observedComposite.Resource.GetAnnotations()[annotationKeyInMemoryEnvEnabled]; inMemoryEnvEnabled != "true" || !found {
		f.log.Debug("In-memory environment config not enabled")
		return rsp, nil
	}

	//  Get the desired composed resources from the request.
	desiredComposed, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}
	f.log.Debug("Found desired resources", "count", len(desiredComposed))

	inMemoryEnvRaw, ok := request.GetContextKey(req, functionContextKeyEnvironment)
	if ok {
		inputEnv := &unstructured.Unstructured{}
		if err := resource.AsObject(inMemoryEnvRaw.GetStructValue(), inputEnv); err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot get Composition environment from %T context key %q", req, functionContextKeyEnvironment))
			return rsp, nil
		}
		f.log.Debug("Loaded Composition environment from Function context", "context-key", functionContextKeyEnvironment)
	}

	envConfig := f.createEnvConfig(inMemoryEnvRaw.GetStructValue().AsMap())

	envConfigLabels := make(map[string]string)
	envConfigLabels["xr-apiversion"] = strings.ReplaceAll(observedComposite.Resource.GetAPIVersion(), "/", "_")
	envConfigLabels["xr-kind"] = observedComposite.Resource.GetKind()
	envConfigLabels["xr-name"] = observedComposite.Resource.GetName()

	envConfig.SetLabels(envConfigLabels)

	_ = xpv1alpha1.AddToScheme(composed.Scheme)
	desiredEnvConfig, err := composed.From(envConfig)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot convert %T to %T", envConfig, &composed.Unstructured{}))
		return rsp, nil
	}

	desiredComposed[("in-memory-env")] = &resource.DesiredComposed{Resource: desiredEnvConfig}
	desiredComposed[("in-memory-env")].Ready = resource.ReadyTrue

	f.log.Info("constructed in-memory EnvConfig", "inMemoryEnv:", desiredEnvConfig)

	if err := response.SetDesiredComposedResources(rsp, desiredComposed); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources from %T", req))
		return rsp, nil
	}

	return rsp, nil
}

func (f *Function) createEnvConfig(data map[string]interface{}) *xpv1alpha1.EnvironmentConfig {
	envConfig := &xpv1alpha1.EnvironmentConfig{}
	envConfig.Data = make(map[string]extv1.JSON)

	keysToSkip := []string{"kind", "apiVersion"}
	for key, value := range data {
		if slices.Contains(keysToSkip, key) {
			continue
		}

		jsonBytes, _ := json.Marshal(value)
		envConfig.Data[key] = extv1.JSON{Raw: jsonBytes}
	}

	return envConfig
}
