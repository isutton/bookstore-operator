# bookstore-operator

The bookstore-operator orchestrates tenants of our BookStore SaaS where each partner has a) an
instance of the BookStore software and b) an instance of our database of choice.

The `bookstore-operator` is meant to be used by Livreiro's operations, an online business that offers
a catalog management system for book stores.

**Note:** This project is for educational purposes only, part of Red Hat OpenShift ShiftWeek's program.

## Implementation

Each tenant is represented in the OpenShift cluster as a `Bookstore` resource, added either through
`kubectl apply -f` or another automated process.

When a new `Bookstore` resource is created, the following operations are executed:

1. A new database resource is created, with its owner reference set to the `Bookstore` resource.
2. A new deployment resource is created, also with its owner reference set to the `Bookstore`
   resource.
3. A new service binding resource is created to enable the `Bookstore` deployment to communicate with
   its respective database resource.

### Example

Below there's a `Bookstore` example (can also be found in `tests/penguin-books.yaml`):

``` yaml
apiVersion: bookstore.livreiro/v1alpha1
kind: Bookstore
metadata:
  name: penguin-books
  namespace: prod
spec:
  customerId: f55e8b5b-05a9-45a9-85c1-f6fe6679a43e
```

To verify no releases are found in the `prod` namespace, use `helm list`:

``` shell
$ helm list -n prod
NAME	NAMESPACE	REVISION	UPDATED	STATUS	CHART	APP VERSION
```

Next, apply the example `Bookstore` resource:

``` shell
$ kubectl apply -f tests/penguin-books.yaml 
bookstore.bookstore.livreiro/penguin-books created
```

The release for `penguin-books`, if everything worked as expected, is now present in `helm list`'s
output:

``` shell
$ helm list -n prod
NAME         	NAMESPACE	REVISION	UPDATED                                 	STATUS  	CHART               	APP VERSION
penguin-books	prod     	1       	2022-03-30 11:16:30.078873078 +0200 CEST	deployed	bookstore-saas-0.1.1	1.16.1
```

The resources specified in the `bookstore-saas` Helm chart should be installed in the cluster.

``` shell
$ kubectl -n prod get all 
NAME                                                READY   STATUS    RESTARTS   AGE
pod/penguin-books-bookstore-saas-6577c4fd8d-hprlq   1/1     Running   0          9m24s

NAME                                   TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
service/penguin-books-bookstore-saas   ClusterIP   10.98.0.139   <none>        80/TCP    9m24s

NAME                                           READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/penguin-books-bookstore-saas   1/1     1            1           9m24s

NAME                                                      DESIRED   CURRENT   READY   AGE
replicaset.apps/penguin-books-bookstore-saas-6577c4fd8d   1         1         1       9m24s
```
