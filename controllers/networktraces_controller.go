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
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	k8sdebuggerv1 "shuanglu/k8s-debugger/api/v1"

	kapp "k8s.io/api/apps/v1"
	kbatch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"

	types "k8s.io/apimachinery/pkg/types"
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
	var networktrace k8sdebuggerv1.Networktrace
	if err := r.Get(ctx, req.NamespacedName, &networktrace); err != nil {
		log.Log.Error(err, "unable to fetch Networktraces")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	//var captureJobs []kbatch.Job
	labels := make(client.MatchingLabels)
	targetPodsMapping := make(map[string][]corev1.Pod)
	for k, v := range networktrace.Spec.Labels {
		labels[k] = v
	}
	var podList corev1.PodList
	r.List(ctx, &podList, client.InNamespace(networktrace.Spec.Namespace))

	switch strings.ToLower(networktrace.Spec.Kind) {
	case "deployment":
		var deploymentList kapp.DeploymentList
		r.List(ctx, &deploymentList, labels)
		// ideally we should only have one deployment
		for _, deployment := range deploymentList.Items {
			var replicasets kapp.ReplicaSetList
			r.List(ctx, &replicasets)
			for _, replicaset := range replicasets.Items {
				for _, rsOwnerReference := range replicaset.OwnerReferences {
					if rsOwnerReference.UID == deployment.ObjectMeta.UID {
						for _, pod := range podList.Items {
							if validatePodOwnerReference(pod, replicaset.GetUID()) {
								targetPodsMapping[pod.Spec.NodeName] = append(targetPodsMapping[pod.Spec.NodeName], pod)
							}
						}
					}
				}
			}
		}

	case "damemonset":
		var daemonSetList kapp.DaemonSetList
		r.List(ctx, &daemonSetList, labels)
		for _, daemonset := range daemonSetList.Items {
			for _, pod := range podList.Items {
				if validatePodOwnerReference(pod, daemonset.GetUID()) {
					targetPodsMapping[pod.Spec.NodeName] = append(targetPodsMapping[pod.Spec.NodeName], pod)
				}
			}
		}

	case "statefulset":
		var statefulsetList kapp.StatefulSetList
		r.List(ctx, &statefulsetList, labels)
		for _, statestatefulset := range statefulsetList.Items {
			for _, pod := range podList.Items {
				if validatePodOwnerReference(pod, statestatefulset.GetUID()) {
					targetPodsMapping[pod.Spec.NodeName] = append(targetPodsMapping[pod.Spec.NodeName], pod)
				}
			}
		}

	case "pod":
		for _, pod := range podList.Items {
			target := true
			for label, value := range labels {
				if v, ok := pod.Labels[label]; ok {
					if v == value {
						continue
					}
				}
				target = false
				break
			}
			if target {
				targetPodsMapping[pod.Spec.NodeName] = append(targetPodsMapping[pod.Spec.NodeName], pod)
			}
		}
	}

	// Fetch the jobs
	var childJobs kbatch.JobList
	if err := r.List(ctx, &childJobs, client.InNamespace(req.Namespace), client.MatchingFields{jobOwnerKey: req.Name}); err != nil {
		log.Log.Error(err, "unable to list Jobs")
		return ctrl.Result{}, err
	}
	if len(childJobs.Items) == len(targetPodsMapping) {
		klog.Infof("Job count match the target count. Will not create new job")
		r.jobStatusCheck(ctx, childJobs, &networktrace)
	} else {
		childJobNode := make(map[string]bool)
		for _, childJob := range childJobs.Items {
			childJobNode[strings.Split(childJob.Name, "_")[1]] = true
		}
		pendingPodsMapping := make(map[string][]corev1.Pod)
		for node, pods := range targetPodsMapping {
			if _, ok := childJobNode[node]; !ok {
				pendingPodsMapping[node] = pods
				tmpJob := &kbatch.Job{}
				tmpJob.Name = "networktrace" + "_" + node
				tmpJob.Spec = kbatch.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							NodeName: node,
						},
					},
				}

			}
		}
	}
	if err := r.Status().Update(ctx, &networktrace); err != nil {
		log.Log.Error(err, "unable to update networktraces status")
	}

	return ctrl.Result{}, nil
}

func (r *NetworktracesReconciler) jobStatusCheck(ctx context.Context, childJobs kbatch.JobList, networktrace *k8sdebuggerv1.Networktrace) {

	// find the active list of jobs
	isJobFinished := func(job *kbatch.Job) (bool, kbatch.JobConditionType) {
		for _, c := range job.Status.Conditions {
			if (c.Type == kbatch.JobComplete || c.Type == kbatch.JobFailed) && c.Status == corev1.ConditionTrue {
				return true, c.Type
			}
		}

		return false, ""
	}

	var activeJobNames []string
	var successfulJobs []*kbatch.Job
	var failedJobs []*kbatch.Job
	for _, childJob := range childJobs.Items {
		finish, finishType := isJobFinished(&childJob)
		if !finish {
			activeJobNames = append(activeJobNames, childJob.Namespace+"_"+childJob.Name)
			continue
		}
		switch finishType {
		case kbatch.JobComplete:
			successfulJobs = append(successfulJobs, &childJob)
		case kbatch.JobFailed:
			failedJobs = append(failedJobs, &childJob)
		}
	}
	if len(activeJobNames) == 0 && (len(successfulJobs) != 0 || len(failedJobs) != 0) {
		networktrace.Status.Completed = "true"
		// Update the failure reason
		for _, job := range failedJobs {
			var pods corev1.PodList
			selectors := job.Spec.Selector.MatchLabels
			labels := make(client.MatchingLabels)
			for k, v := range selectors {
				labels[k] = v
			}
			err := r.List(ctx, &pods, client.InNamespace(networktrace.ObjectMeta.Namespace), labels)
			if err != nil {
				log.Log.Error(err, "unable to list Pods")
			}
			// ideally we should only have one pod
			for _, pod := range pods.Items {
				for _, containerStatus := range pod.Status.ContainerStatuses {
					if containerStatus.Name == "networktracesagent" {
						networktrace.Status.Failed[job.Name] = containerStatus.State.Terminated.Message
					}
				}
			}
		}
	} else {
		networktrace.Status.Completed = "false"
		networktrace.Status.Active = activeJobNames
	}

}

func validatePodOwnerReference(pod corev1.Pod, uid types.UID) bool {
	for _, ownerReference := range pod.OwnerReferences {
		if ownerReference.UID == uid {
			return true
		}
	}
	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *NetworktracesReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&k8sdebuggerv1.Networktrace{}).
		Complete(r)
}
