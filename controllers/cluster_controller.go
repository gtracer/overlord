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

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/gtracer/overlord/api/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Log logr.Logger
}

// +kubebuilder:rbac:groups=kubernetes.ov3rlord.me,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubernetes.ov3rlord.me,resources=clusters/status,verbs=get;update;patch

func (r *ClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("cluster", req.NamespacedName)

	cluster := &v1.Cluster{}

	err := r.Get(ctx, req.NamespacedName, cluster)
	if err != nil {
		log.Error(err, "error get cluster")
		return ctrl.Result{}, err
	}

	minionList := &v1.MinionList{}
	matchingLabels := client.MatchingLabels(
		map[string]string{
			"kubernetes.ov3rlord.me/cluster": req.Name,
		})
	err = r.List(ctx, minionList, client.InNamespace(req.Namespace), matchingLabels)
	if err != nil {
		log.Error(err, "error listing minions")
		return ctrl.Result{}, err
	}

	for _, minion := range minionList.Items {
		if minion.Spec.Master != minion.Name {
			continue
		}
		cluster.Status.Kubeconfig = minion.Status.Kubeconfig
		cluster.Status.Token = minion.Status.Token
		cluster.Status.Master = minion.Name
	}

	if cluster.Status.Master == "" &&
		len(minionList.Items) > 0 {
		cluster.Status.Master = minionList.Items[0].Name
	}
	err = r.Status().Update(ctx, cluster)
	if err != nil {
		log.Error(err, "error update cluster status")
		return ctrl.Result{}, err
	}

	for _, minion := range minionList.Items {
		if minion.Spec.Master == cluster.Status.Master {
			continue
		}
		minion.Spec.Master = cluster.Status.Master
		err = r.Update(ctx, &minion)
		if err != nil {
			log.Error(err, "error update cluster status")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.Cluster{}).
		Complete(r)
}
