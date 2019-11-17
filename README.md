# secret-replicator [![Build Status](https://travis-ci.org/thomas-maurice/secret-replicator.svg?branch=master)](https://travis-ci.org/thomas-maurice/secret-replicator)

This is a simple controller to copy secrets from a namespace to another and keep them in sync.

## Installation

```
kubectl apply -f https://raw.githubusercontent.com/thomas-maurice/secret-replicator/master/dist/dist.yaml
```

## Adding a replication
Create an object as follows:

```
apiVersion: replication.apis.maurice.fr/v1
kind: SecretReplication
metadata:
  name: example-replication
spec:
  srcNamespace: default
  dstNamespace: test-replication
  srcName: secret
  dstName: secret-copy
```

The controller will check the source secret periodically and if its version has changed, will mirror the changes to the destination one.