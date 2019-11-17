# secret-replicator [![Build Status](https://travis-ci.org/thomas-maurice/secret-replicator.svg?branch=master)](https://travis-ci.org/thomas-maurice/secret-replicator)

This is a simple controller to copy secrets from a namespace to another and keep them in sync.

## Installation

```
kubectl apply -f https://raw.githubusercontent.com/thomas-maurice/secret-replicator/master/dist/dist.yaml
```

## Adding a replication

Create a sample secret:
```
$ kubectl create secret generic hello
secret/hello created
```

Create the destination namespace:
```
$ kubectl create namespace test-replication
namespace/test-replication created
```

Create an object as follows:

```
apiVersion: replication.apis.maurice.fr/v1
kind: SecretReplication
metadata:
  name: example-replication
spec:
  srcNamespace: default
  dstNamespace: test-replication
  srcName: hello
  dstName: hello-copy
```

`kubectl apply` it, then check the secret in the destination namespace:
```
$ kubectl apply -f replication.yaml
secretreplication.replication.apis.maurice.fr/example-replication created
$ kubectl get secrets -n test-replication
NAME                  TYPE                                  DATA   AGE
default-token-xprlx   kubernetes.io/service-account-token   3      84s
hello-copy            Opaque                                0      35s
```

The controller will check the source secret periodically and if its version has changed, will mirror the changes to the destination one.