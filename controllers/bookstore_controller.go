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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	bookstorev1beta1 "github.com/isutton/bookstore-operator/api/v1beta1"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/kube"
	"helm.sh/helm/v3/pkg/release"
	"helm.sh/helm/v3/pkg/storage"
	"helm.sh/helm/v3/pkg/storage/driver"
)

// BookstoreReconciler reconciles a Bookstore object
type BookstoreReconciler struct {
	client.Client
	Scheme     *runtime.Scheme
	ChartDir   string
	RESTConfig *rest.Config
}

func NewBookstoreReconciler(mgr ctrl.Manager, restConfig *rest.Config, chartDir string) *BookstoreReconciler {
	return &BookstoreReconciler{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		ChartDir:   chartDir,
		RESTConfig: restConfig,
	}
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
	logger := log.FromContext(ctx)

	coreClient := corev1.NewForConfigOrDie(r.RESTConfig)
	cmDriver := driver.NewSecrets(coreClient.Secrets(req.NamespacedName.Namespace))
	getter := genericclioptions.NewConfigFlags(true)
	getter.Namespace = &req.NamespacedName.Namespace
	kubeClient := kube.New(getter)
	kubeClient.Namespace = req.NamespacedName.Namespace

	config := &action.Configuration{
		Releases:         storage.Init(cmDriver),
		KubeClient:       kubeClient,
		Log:              kubeClient.Log,
		RESTClientGetter: getter,
	}

	if ok, err := r.IsUninstall(ctx, req.NamespacedName); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if ok {
		logger.Info("Uninstalling tenant")
		if uninstallErr := r.Uninstall(config, req.NamespacedName); uninstallErr != nil {
			return ctrl.Result{Requeue: true}, uninstallErr
		}
		return ctrl.Result{}, nil
	}

	if ok, err := r.IsInstall(ctx, config, req.NamespacedName); err != nil {
		return ctrl.Result{Requeue: true}, err
	} else if ok {
		logger.Info("Installing tenant")
		if installErr := r.Install(ctx, config, req.NamespacedName); installErr != nil {
			return ctrl.Result{Requeue: true}, installErr
		}
	} else {
		logger.Info("Upgrading tenant")
		if upgradeErr := r.Upgrade(ctx, config, req.NamespacedName); upgradeErr != nil {
			return ctrl.Result{Requeue: true}, upgradeErr
		}
	}

	return ctrl.Result{}, nil
}

func (r *BookstoreReconciler) GetReleases(config *action.Configuration) ([]*release.Release, error) {
	return action.NewList(config).Run()
}

func (r *BookstoreReconciler) GetBookstore(ctx context.Context, namespacedName types.NamespacedName) (*bookstorev1beta1.Bookstore, error) {
	bookstore := &bookstorev1beta1.Bookstore{}
	if err := r.Client.Get(ctx, namespacedName, bookstore); err != nil {
		return bookstore, err
	}
	return bookstore, nil
}

func (r *BookstoreReconciler) IsUninstall(ctx context.Context, namespacedName types.NamespacedName) (bool, error) {
	_, err := r.GetBookstore(ctx, namespacedName)
	if err != nil {
		if errors.IsNotFound(err) {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

func (r *BookstoreReconciler) IsInstall(ctx context.Context, config *action.Configuration, namespacedName types.NamespacedName) (bool, error) {
	releases, err := r.GetReleases(config)
	if err != nil {
		return false, err
	}

	foundRelease := false
	for _, release := range releases {
		if release.Name == namespacedName.Name &&
			release.Namespace == namespacedName.Namespace {
			foundRelease = true
			break
		}
	}

	return !foundRelease, nil
}

func (r *BookstoreReconciler) Install(ctx context.Context, config *action.Configuration, namespacedName types.NamespacedName) error {
	chrt, loadChartErr := r.LoadChart()
	if loadChartErr != nil {
		return loadChartErr
	}

	i := action.NewInstall(config)
	i.ReleaseName = namespacedName.Name
	i.Namespace = namespacedName.Namespace

	_, installErr := i.RunWithContext(ctx, chrt, nil)

	return installErr
}

func (r *BookstoreReconciler) GetRelease(ctx context.Context, config *action.Configuration, namespacedName types.NamespacedName) (*release.Release, error) {
	releases, err := r.GetReleases(config)
	if err != nil {
		return nil, err
	}

	for _, release := range releases {
		if release.Name == namespacedName.Name && release.Namespace == namespacedName.Namespace {
			return release, nil
		}
	}

	return nil, nil
}

func (r *BookstoreReconciler) Upgrade(ctx context.Context, config *action.Configuration, namespacedName types.NamespacedName) error {
	chrt, loadChartErr := r.LoadChart()
	if loadChartErr != nil {
		return loadChartErr
	}

	release, err := r.GetRelease(ctx, config, namespacedName)
	if err != nil {
		return err
	}

	if release == nil {
		return nil
	}

	if release.Chart.AppVersion() == chrt.AppVersion() {
		return nil
	}

	u := action.NewUpgrade(config)
	u.Namespace = namespacedName.Namespace

	_, upgradeErr := action.NewUpgrade(config).RunWithContext(ctx, namespacedName.Name, chrt, nil)

	return upgradeErr
}

func (r *BookstoreReconciler) Uninstall(config *action.Configuration, namespacedName types.NamespacedName) error {
	_, err := action.NewUninstall(config).Run(namespacedName.Name)
	return err
}

func (r *BookstoreReconciler) LoadChart() (*chart.Chart, error) {
	return loader.LoadDir(r.ChartDir)
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookstoreReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&bookstorev1beta1.Bookstore{}).
		Complete(r)
}
