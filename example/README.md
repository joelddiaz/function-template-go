# Example manifests

You can run your function locally and test it using `crossplane beta render`
with these example manifests.

```shell
# Run the function locally
$ go run . --insecure --debug
```

Or start the debuger from VSCode.

```shell
# Then, in another terminal, call it with these example manifests
# Because we're using the patch-and-transform function in this example, that container
# will be run as a local Docker container to perform the first part of the example composition.
$ crossplane beta render xr.yaml composition.yaml functions.yaml -r
---
apiVersion: apiextensions.crossplane.io/v1
kind: CompositeResourceDefinition
metadata:
  name: nosqls.database.example.com
---
apiVersion: s3.aws.upbound.io/v1beta1
kind: Bucket
metadata:
  annotations:
    crossplane.io/composition-resource-name: s3Bucket
  generateName: nosqls.database.example.com-
  labels:
    crossplane.io/composite: nosqls.database.example.com
  ### Here we see the function changed the name of the Bucket resources
  name: NewNameXYZ
  ownerReferences:
  - apiVersion: apiextensions.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: CompositeResourceDefinition
    name: nosqls.database.example.com
    uid: ""
spec:
  forProvider:
    region: us-east-1
  providerConfigRef:
    name: default
---
apiVersion: dynamodb.aws.upbound.io/v1beta1
kind: Table
metadata:
  annotations:
    crossplane.io/composition-resource-name: dynamoDB
  generateName: nosqls.database.example.com-
  labels:
    crossplane.io/composite: nosqls.database.example.com
  ownerReferences:
  - apiVersion: apiextensions.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: CompositeResourceDefinition
    name: nosqls.database.example.com
    uid: ""
spec:
  forProvider:
    attribute:
    - name: S3ID
      type: S
    hashKey: S3ID
    readCapacity: 1
    region: us-east-1
    writeCapacity: 1
---
### And here we see that the function added a brand new Bucket resources
apiVersion: s3.aws.crossplane.io/v1beta1
kind: Bucket
metadata:
  annotations:
    crossplane.io/composition-resource-name: dynamicXPlaneFnBucket
  generateName: nosqls.database.example.com-
  labels:
    crossplane.io/composite: nosqls.database.example.com
  name: NewXPlaneFnBucket
  ownerReferences:
  - apiVersion: apiextensions.crossplane.io/v1
    blockOwnerDeletion: true
    controller: true
    kind: CompositeResourceDefinition
    name: nosqls.database.example.com
    uid: ""
spec:
  forProvider:
    locationConstraint: ""
status:
  atProvider:
    arn: ""
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
message: I was run with input :{"ExtraBucket"}!
severity: SEVERITY_NORMAL
step: my-step
```
