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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sdebuggerv1 "shuanglu/k8s-debugger/api/v1"

	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

var (
	jobOwnerKey = ".metadata.controller"
)

// NetworktracesReconciler reconciles a Networktraces object
type NetworktracesReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=k8s-debugger.shuanglu.io,resources=networktraces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=k8s-debugger.shuanglu.io,resources=networktraces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=k8s-debugger.shuanglu.io,resources=networktraces/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Networktraces object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *NetworktracesReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	// Fetch the networktraces object
	var networktraces k8sdebuggerv1.Networktraces
	if err := r.Get(ctx, req.NamespacedName, &networktraces); err != nil {
		log.Log.Error(err, "unable to fetch Networktraces")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch the pods

	// Fetch the jobs
	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Log.Error(err, "unable to list Jobs")
		return ctrl.Result{}, err
	}

	// find the active list of jobs
	isJobFinished := func(job *kbatch.Job) (bool, kbatch.JobConditionType) {
		for _, c := range job.Status.Conditions {
			if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
				return true, c.Type
			}
		}

		return false, ""
	}

	var activeJobs []*kbatch.Job
	var successfulJobs []*kbatch.Job
	var failedJobs []*kbatch.Job
	for _, childJob := range childJobs.Items {
		_, finishType := isJobFinished(&childJob)
		switch finishType {
		case "":
			activeJobs = append(activeJobs, &childJob)
		case kbatch.JobComplete:
			successfulJobs = append(successfulJobs, &childJob)
		case kbatch.JobFailed:
			failedJobs = append(failedJobs, &childJob)
		}
	}
	if len(activeJobs) == 0 && (len(successfulJobs) != 0 || len(failedJobs) != 0) {
		networktraces.Status.Completed = "true"
	}

	// Update the failure reason
	for _, job := range failedJobs {
		var pods corev1.PodList
		selectors := job.Spec.Selector.MatchLabels
		labels := make(client.MatchingLabels)
		for k, v := range selectors {
			labels[k] = v
		}
		err := r.List(ctx, &pods, client.InNamespace(req.Namespace), labels)
		if err != nil {
			log.Log.Error(err, "unable to list Pods")
		}
		// ideally we should only have one pod
		for _, pod := range pods.Items {
			for _, containerStatus := range pod.Status.ContainerStatuses {
				if containerStatus.Name == "networktracesagent" {
					networktraces.Status.Failed[job.Name] = containerStatus.State.Terminated.Message
				}
			}
		}
	}
	if err := r.Status().Update(ctx, &networktraces); err != nil {
		log.Log.Error(err, "unable to update networktraces status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetworktracesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sdebuggerv1.Networktraces{}).
		Complete(r)
}
