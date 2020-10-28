/*


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
	"time"

	"github.com/hairizuanbinnoorazman/sheetops/appmgr"
	"github.com/hairizuanbinnoorazman/sheetops/logger"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	sheetopsv1alpha1 "github.com/hairizuanbinnoorazman/sheetops/api/v1alpha1"
)

// To prevent crazy number of calls
var appSettings []appmgr.AppSetting
var lastAPICall time.Time

// GooglesheetSyncReconciler reconciles a GooglesheetSync object
type GooglesheetSyncReconciler struct {
	client.Client
	Log      logger.Logger
	Scheme   *runtime.Scheme
	SheetSvc appmgr.GoogleSheets
}

// +kubebuilder:rbac:groups=sheetops.hairizuan.com,resources=googlesheetsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sheetops.hairizuan.com,resources=googlesheetsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;create;update;delete;watch

func (r *GooglesheetSyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	r.Log.Info("Begin reconcilation for sheetops - googlesheetsync resource")
	defer r.Log.Info("End reconcilation for sheetops - googlesheetsync resource")

	var syncer sheetopsv1alpha1.GooglesheetSync
	err := r.Get(ctx, req.NamespacedName, &syncer)

	if err != nil {
		r.Log.Errorf("Unable to retrieve item from kubernetes api. Will retry again soon. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 60 * time.Second,
		}, nil
	}

	syncer.Status.SyncStatus = "reconciling"

	err = r.Status().Update(ctx, &syncer)
	if err != nil {
		r.Log.Errorf("Unable to update status of syncer - indicative of more issues for the integration. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 60 * time.Second,
		}, nil
	}

	if time.Now().Sub(lastAPICall).Seconds() > float64(60) {
		r.Log.Infof("Begin reconcilation of app settings from %v - cellrange: %v", syncer.Spec.SpreadsheetID, syncer.Spec.CellRange)
		appSettings, err = r.SheetSvc.GetValues(syncer.Spec.SpreadsheetID, syncer.Spec.CellRange)
		lastAPICall = time.Now()
	} else {
		r.Log.Infof("Utilize cached version of app settings. AppSettings: %+v", appSettings)
	}

	if err != nil {
		r.Log.Errorf("Unable to retrieve app settings from googlesheets. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 60 * time.Second,
		}, nil
	}

	errorsFound := false
	for _, appSetting := range appSettings {
		r.Log.Infof("Attempting to reconcile %v", appSetting.Name)
		var expectedDeployment appsv1.Deployment
		err = r.Get(ctx, types.NamespacedName{
			Namespace: "default",
			Name:      appSetting.Name}, &expectedDeployment)

		// Assume that deployment is not found
		if err != nil {
			r.Log.Infof("Deployment %v is not found. We will attempt to create it", appSetting.Name)
			var newDeployment appsv1.Deployment
			newDeployment.SetName(appSetting.Name)
			newDeployment.SetNamespace("default")
			newDeployment.SetLabels(map[string]string{
				"managed-by": "googlespreadsheet-sync",
				"last-sync":  time.Now().Format("2006-01-02T1504"),
			})
			ls := &meta.LabelSelector{}
			ls = meta.AddLabelToSelector(ls, "app", appSetting.Name)
			newDeployment.Spec.Selector = ls
			newDeployment.Spec.Replicas = int32convert(appSetting.Replicas)
			cont := v1.Container{
				Name:  appSetting.Name,
				Image: appSetting.Image,
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("300m"),
						v1.ResourceMemory: resource.MustParse("300Mi"),
					},
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("100m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
					},
				},
			}
			newDeployment.Spec.Template.ObjectMeta.Labels = map[string]string{
				"app": appSetting.Name,
			}
			newDeployment.Spec.Template.Spec.Containers = []v1.Container{cont}
			err = r.Create(ctx, &newDeployment)
			if err != nil {
				r.Log.Errorf("Unable to create the new deployment. Err: %v", err)
				errorsFound = true
			}
			continue
		}

		// Assume deployment is found
		if expectedDeployment.Spec.Template.Spec.Containers[0].Image != appSetting.Image || *expectedDeployment.Spec.Replicas != int32(appSetting.Replicas) {
			expectedDeployment.SetLabels(map[string]string{
				"managed-by": "googlespreadsheet-sync",
				"last-sync":  time.Now().Format("2006-01-02T1504"),
			})
			expectedDeployment.Spec.Template.Spec.Containers[0].Image = appSetting.Image
			expectedDeployment.Spec.Replicas = int32convert(appSetting.Replicas)
			err = r.Update(ctx, &expectedDeployment)
			if err != nil {
				r.Log.Errorf("Unable to update the deployment. Err: %v", err)
				errorsFound = true
			}
			continue
		}

		r.Log.Infof("Application %v on cluster is as expected. Will proceed to next one", appSetting.Name)
		continue
	}

	if errorsFound {
		syncer.Status.SyncStatus = "reconcile errors"
		err = r.Status().Update(ctx, &syncer)
		if err != nil {
			r.Log.Errorf("Unable to update status of syncer - indicative of more issues for the integration. Err: %v", err)
			return ctrl.Result{
				RequeueAfter: 60 * time.Second,
			}, nil
		}
	}

	syncer.Status.SyncStatus = "reconcile complete"
	err = r.Status().Update(ctx, &syncer)
	if err != nil {
		r.Log.Errorf("Unable to update status of syncer - indicative of more issues for the integration. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 60 * time.Second,
		}, nil
	}

	return ctrl.Result{
		RequeueAfter: 60 * time.Second,
	}, nil
}

func int32convert(i int) *int32 {
	j := int32(i)
	return &j
}

func (r *GooglesheetSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sheetopsv1alpha1.GooglesheetSync{}).
		WithEventFilter(predicate.GenerationChangedPredicate{}).
		Complete(r)
}
