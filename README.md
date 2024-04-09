# dynamic-refresh-operator
dynamically watch secret or configmap and refresh deployment 

## Build image
```
./hack/build.sh
```

## Run Locally
```
./hack/run-locally.sh
```

## Deploy in K8s
```
./hack/kube-deploy.sh
```

## Build Bundle for OLM
```
cd ./hack
./create-bundle
```

```
# apply generated yaml for CatalogSource
# make sure the index and bundle images repo are public

# Verify
oc get catalogsources -A
oc get pods -n openshift-marketplace -w
oc get packagemanifest -A |  grep dynamic
```