/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bookstorev1beta1 "github.com/isutton/bookstore-operator/api/v1beta1"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
)

// BookstoreReconciler reconciles a Bookstore object
type BookstoreReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	RESTConfig *rest.Config
}

//+kubebuilder:rbac:groups=bookstore.livreiro,resources=bookstores,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=bookstore.livreiro,resources=bookstores/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=bookstore.livreiro,resources=bookstores/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Bookstore object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *BookstoreReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO: find out how to populate a RESTClientGetter from the
	//       information we have at hand (for example the
	//       r.RESTConfig field recently added).
	kubeClient := kube.New(nil)

	// TODO: share r.Client with cfg below, so `listAction.Run()`
	//       benefits from the same cache.
	cfg := &action.Configuration{
		KubeClient: kubeClient,
		Log:        kubeClient.Log,
	}

	listAction := action.NewList(cfg)
	releases, err := listAction.Run()
	if err != nil {
		// errors here are exceptional as it only requires a
		// read-only access to the cluster to list releases.
		return ctrl.Result{Requeue: true}, err
	}

	// TODO: simplify this flow moving code into smaller
	//       functions; the flow here doesn't read clear as I
	//       mixed both and non-functional flows.
	bookstore := &bookstorev1beta1.Bookstore{}
	if err := r.Client.Get(ctx, req.NamespacedName, bookstore); err != nil {
		if !errors.IsNotFound(err) {
			return ctrl.Result{Requeue: true}, err
		}
		uninstallAction := action.NewUninstall(cfg)
		_, err := uninstallAction.Run(req.NamespacedName.Name)
		if err != nil {
			return ctrl.Result{Requeue: true}, err
		}
		return ctrl.Result{Requeue: false}, nil
	}

	// TODO: embed chart bytes instead.
	chrt, err := loader.LoadDir(".")
	if err != nil {
		return ctrl.Result{Requeue: false}, err
	}

	// process to check whether releases do exist for the current bookstore.
	foundRelease := false
	for _, release := range releases {
		if release.Name == req.NamespacedName.Name &&
			release.Namespace == req.NamespacedName.Namespace {
			foundRelease = true
			break
		}
	}

	if !foundRelease {
		installAction := action.NewInstall(cfg)
		_, err := installAction.Run(chrt, nil)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}
	} else {
		updateAction := action.NewUpgrade(cfg)
		_, err := updateAction.Run(req.NamespacedName.Name, chrt, nil)
		if err != nil {
			return ctrl.Result{Requeue: false}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookstoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bookstorev1beta1.Bookstore{}).
		Complete(r)
}
