package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

func TestRunFunction(t *testing.T) {

	type args struct {
		ctx context.Context
		req *fnv1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason              string
		args                args
		want                want
		desiredResourcesLen int
	}{
		"ResponseIsReturnedWithoutS3BucketResource": {
			reason: "The Function should return a fatal result if no input was specified",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta: &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: resource.MustStructJSON(`{
						"apiVersion": "template.fn.crossplane.io/v1beta1",
						"kind": "Input",
						"extras": {
							"tgwMode": "NoExtraBucket"
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1beta1.Result{
						{
							Severity: fnv1beta1.Severity_SEVERITY_NORMAL,
							Message:  "I was run with input :{\"NoExtraBucket\"}!",
						},
					},
				},
			},
			desiredResourcesLen: 0,
		},
		"ResponseIncludesAddedS3BucketResource": {
			reason: "The Function should add an S3 bucket if input is set to ExtraBucket",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta: &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: resource.MustStructJSON(`{
						"apiVersion": "template.fn.crossplane.io/v1beta1",
						"kind": "Input",
						"extras": {
							"tgwMode": "ExtraBucket"
						}
					}`),
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1beta1.Result{
						{
							Severity: fnv1beta1.Severity_SEVERITY_NORMAL,
							Message:  "I was run with input :{\"ExtraBucket\"}!",
						},
					},
				},
			},
			desiredResourcesLen: 1,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)
			if err != nil {
				t.Errorf("RunFunction() returned an error")
				t.Fail()
			}
			// check Severity
			if diff := cmp.Diff(tc.want.rsp.Results[0].Severity, rsp.Results[0].Severity, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want severity, +got severity:\n%s", tc.reason, diff)
			}
			if diff := cmp.Diff(tc.want.rsp.Results[0].Message, rsp.Results[0].Message); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): - want message, +got message:\n%s", tc.reason, diff)
			}

			if tc.desiredResourcesLen != len(rsp.Desired.Resources) {
				t.Errorf("Expected number of returned resources to be %d, got %d", tc.desiredResourcesLen, len(rsp.Results))
			}
		})
	}
}
