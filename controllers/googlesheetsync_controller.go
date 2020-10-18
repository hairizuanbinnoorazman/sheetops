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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	sheetopsv1alpha1 "github.com/hairizuanbinnoorazman/sheetops/api/v1alpha1"
)

// GooglesheetSyncReconciler reconciles a GooglesheetSync object
type GooglesheetSyncReconciler struct {
	client.Client
	Log      logger.Logger
	Scheme   *runtime.Scheme
	SheetSvc appmgr.GoogleSheets
}

// +kubebuilder:rbac:groups=sheetops.hairizuan.com,resources=googlesheetsyncs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sheetops.hairizuan.com,resources=googlesheetsyncs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployment,verbs=get;list;create;update;delete

func (r *GooglesheetSyncReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()

	r.Log.Info("Begin reconcilation for sheetops - googlesheetsync resource")
	defer r.Log.Info("End reconcilation for sheetops - googlesheetsync resource")

	var syncer sheetopsv1alpha1.GooglesheetSync
	err := r.Get(ctx, req.NamespacedName, &syncer)

	if err != nil {
		r.Log.Errorf("Unable to retrieve item from kubernetes api. Will retry again soon. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	syncer.Status.SyncStatus = "reconciling"

	err = r.Status().Update(ctx, &syncer)
	if err != nil {
		r.Log.Errorf("Unable to update status of syncer - indicative of more issues for the integration. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	r.Log.Infof("Begin reconcilation of app settings from %v - cellrange: %v", syncer.Spec.SpreadsheetID, syncer.Spec.CellRange)
	appSettings, err := r.SheetSvc.GetValues(syncer.Spec.SpreadsheetID, syncer.Spec.CellRange)

	if err != nil {
		r.Log.Errorf("Unable to retrieve app settings from googlesheets. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
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
				"last-sync":  time.Now().String(),
			})
			newDeployment.Spec.Replicas = int32convert(appSetting.Replicas)
			newDeployment.Spec.Template.Spec.Containers[0].Image = appSetting.Image
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
				"last-sync":  time.Now().String(),
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
				RequeueAfter: 10 * time.Second,
			}, nil
		}
	}

	syncer.Status.SyncStatus = "reconcile complete"
	err = r.Status().Update(ctx, &syncer)
	if err != nil {
		r.Log.Errorf("Unable to update status of syncer - indicative of more issues for the integration. Err: %v", err)
		return ctrl.Result{
			RequeueAfter: 10 * time.Second,
		}, nil
	}

	return ctrl.Result{
		RequeueAfter: 10 * time.Second,
	}, nil
}

func int32convert(i int) *int32 {
	j := int32(i)
	return &j
}

func (r *GooglesheetSyncReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sheetopsv1alpha1.GooglesheetSync{}).
		Complete(r)
}
