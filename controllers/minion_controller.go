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
	"fmt"
	"time"

	"github.com/go-logr/logr"
	v1 "github.com/gtracer/overlord/api/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MinionReconciler reconciles a Minion object
type MinionReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=kubernetes.ov3rlord.me,resources=minions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubernetes.ov3rlord.me,resources=minions/status,verbs=get;update;patch

func (r *MinionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("minion", req.NamespacedName)

	minion := &v1.Minion{}
	err := r.Get(ctx, req.NamespacedName, minion)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "failed to get minion")
		return ctrl.Result{}, err
	}

	if time.Since(minion.Status.LastTimestamp.Time).Seconds() > 120 {
		minion.Status.NodeStatus.State = v1.Unhealthy
		minion.Status.NodeStatus.Message =
			fmt.Sprintf("no healthy report since %s", minion.Status.LastTimestamp.Time.String())
		err = r.Status().Update(ctx, minion)
		if err != nil {
			log.Error(err, "failed to update minion status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *MinionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Minion{}).
		Complete(r)
}
