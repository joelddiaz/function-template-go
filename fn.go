package main

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/resource/composed"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/crossplane/function-template-go/input/v1beta1"

	s3v1beta1 "github.com/crossplane-contrib/provider-aws/apis/s3/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

func init() {
	s3v1beta1.SchemeBuilder.AddToScheme(composed.Scheme)
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	f.log.Debug("Received input:", "in", in)

	// The composite resource that actually exists.
	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get observed composite resource"))
		return rsp, nil
	}

	f.log.Debug("Observed composite resource", "oxr", oxr)

	dxr, err := request.GetDesiredCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composite resources"))
		return rsp, nil
	} else {
		f.log.Debug("Desired composite resource", "dxr", dxr)
	}

	dcr, err := request.GetDesiredComposedResources(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	} else {
		f.log.Info("Desired COMPOSED resources", "dcr", dcr)
	}

	// Create a mapping to hold the resources to be returned when
	// our function has completed.
	newDesired := map[resource.Name]*resource.DesiredComposed{}

	// Iterate through list of composed resources and in this example we
	// will change the name of any Bucket resources we come across
	for name, desiredRes := range dcr {
		f.log.Debug("Checking on resource...", "key", name)
		if desiredRes.Resource.GetKind() == "Bucket" {
			f.log.Debug("Changing name on resource type of Bucket")
			desiredRes.Resource.SetName("NewNameXYZ")
		}

		newDesired[name] = desiredRes
	}

	// if the input is as expected, create and add a new S3 bucket to the map of resources
	if in.Extras.ExampleFlag == "ExtraBucket" {
		newBucket := &s3v1beta1.Bucket{
			ObjectMeta: metav1.ObjectMeta{
				Name: "NewXPlaneFnBucket",
			},
			Spec: s3v1beta1.BucketSpec{
				ForProvider: s3v1beta1.BucketParameters{},
			},
		}

		newRes := resource.NewDesiredComposed()

		newRes.Resource, err = composed.From(newBucket)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "failed to create new Bucket resource from %T", newBucket))
			return rsp, nil
		}

		newDesired["dynamicXPlaneFnBucket"] = newRes
	}

	f.log.Debug("List of returned resources:")
	for k, v := range newDesired {
		f.log.Debug("Desired resource:", "name", k, "val", v)
	}

	// Send the map of new desired resources in our response.
	if err := response.SetDesiredComposedResources(rsp, newDesired); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composed resources from %T", req))
		return rsp, nil
	}

	response.Normalf(rsp, "I was run with input :%q!", in.Extras)

	return rsp, nil
}
