// /*
// Copyright 2023.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

//     http://www.apache.org/licenses/LICENSE-2.0

// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// */

// Create: The Operator will act on the CR to create a CronJob that will run the health check. The operator should create a CronJob when CR is created. The operator should monitor the namespace it's running in for the CR. The operator should not monitor ALL namespaces for CR. CRs created in other namespaces will be ignored. The status & details such as last execution time & status of the CronJob should be added to the CR at each reconciliation.
// Update: The operator should update the CronJob when CR is updated. For example, if user updates health check's frequency, the CronJob schedule should be updated as well.
// Delete: The operator should delete the CronJob when CR is deleted. Operator should add finalizers to all resources it creates. CR should also be deleted once all child resources such as CronJob, ConfigMap, etc. are deleted.

package controller

import (
	"context"
	"fmt"
	"reflect"
	"strconv"

	logs "github.com/go-logr/logr"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import your CRD package

	monitoringv1alpha1 "httpcheck.io/api/v1alpha1"
)

// CronJobReconciler reconciles a CronJob object
type CronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logs.Logger
}

func (r *CronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cronjob", req.NamespacedName)

	cronJob := &monitoringv1alpha1.CronJob{}
	err := r.Get(ctx, req.NamespacedName, cronJob)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get CronJob")
		return ctrl.Result{}, err
	}

	if cronJob.DeletionTimestamp != nil {
		return ctrl.Result{}, nil
	}

	builtCronJob, err := r.buildCronJob(cronJob)
	if err != nil {
		log.Error(err, "Failed to build CronJob")
		return ctrl.Result{}, err
	}

	existingCronJob := &batchv1.CronJob{}
	err = r.Get(ctx, types.NamespacedName{Name: builtCronJob.Name, Namespace: builtCronJob.Namespace}, existingCronJob)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(ctx, builtCronJob)
		if err != nil {
			log.Error(err, "Failed to create CronJob")
			return ctrl.Result{}, err
		}
		log.Info("CronJob created successfully")
	} else if err != nil {
		log.Error(err, "Failed to get existing CronJob")
		return ctrl.Result{}, err
	} else {
		// Update the existing CronJob if it's different from the built one
		if !reflect.DeepEqual(existingCronJob.Spec, builtCronJob.Spec) {
			existingCronJob.Spec = builtCronJob.Spec
			err = r.Update(ctx, existingCronJob)
			if err != nil {
				log.Error(err, "Failed to update CronJob")
				return ctrl.Result{}, err
			}
			log.Info("CronJob updated successfully")
		}
	}
	lastExecutionTime := metav1.Now()

	// Update the CronJob status
	if !cronJob.Status.LastExecutionTime.Time.Equal(lastExecutionTime.Time) {
		cronJob.Status.LastExecutionTime = metav1.Time{Time: lastExecutionTime.Time}
		err := r.Status().Update(ctx, cronJob)
		if err != nil {
			log.Error(err, "Failed to update CronJob status")
			return ctrl.Result{}, err
		}
		log.Info("CronJob status updated successfully")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitoringv1alpha1.CronJob{}).
		Owns(&corev1.Pod{}).
		// Watches(&source.Kind{Type: &monitoringv1alpha1.CronJob{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
func (r *CronJobReconciler) buildCronJob(cronJob *monitoringv1alpha1.CronJob) (*batchv1.CronJob, error) {
	labels := map[string]string{
		"app": cronJob.Name,
	}
	schedule, _ := calculateCronSchedule(cronJob.Spec.CheckIntervalInSeconds)

	envVars := []corev1.EnvVar{
		{Name: "KAFKA_BOOTSTRAP_SERVERS", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "kafka-config"},
				Key:                  "KAFKA_BOOTSTRAP_SERVERS",
			},
		}},
		{Name: "KAFKA_TOPIC", ValueFrom: &corev1.EnvVarSource{
			ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{Name: "kafka-config"},
				Key:                  "KAFKA_TOPIC",
			},
		}},
		{Name: "ID", Value: cronJob.Spec.ID},
		{Name: "NAME", Value: cronJob.Spec.Name},
		{Name: "URI", Value: cronJob.Spec.URI},
		{Name: "IS_PAUSED", Value: strconv.FormatBool(cronJob.Spec.IsPaused)},
		{Name: "NUM_RETRIES", Value: strconv.Itoa(cronJob.Spec.NumRetries)},
		{Name: "UPTIME_SLA", Value: strconv.Itoa(cronJob.Spec.UptimeSLA)},
		{Name: "RESPONSE_TIME_SLA", Value: strconv.Itoa(cronJob.Spec.ResponseTimeSLA)},
		{Name: "USE_SSL", Value: strconv.FormatBool(cronJob.Spec.UseSSL)},
		{Name: "RESPONSE_STATUS_CODE", Value: strconv.Itoa(cronJob.Spec.ResponseStatusCode)},
		{Name: "CHECK_INTERVAL_IN_SECONDS", Value: strconv.Itoa(cronJob.Spec.CheckIntervalInSeconds)},
	}
	imagePullSecrets := []corev1.LocalObjectReference{
		{Name: "registry-secret"},
	}

	cronJobObject := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cronJob.Name,
			Namespace: cronJob.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.CronJobSpec{
			Schedule: schedule,
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "producer",
									Image: "abhigade123/producer1:latest",
									Env:   envVars,
								},
							},
							ImagePullSecrets: imagePullSecrets,
							RestartPolicy:    corev1.RestartPolicyOnFailure,
						},
					},
				},
			},
		},
	}

	// Set the owner reference
	if err := ctrl.SetControllerReference(cronJob, cronJobObject, r.Scheme); err != nil {
		return nil, err
	}

	return cronJobObject, nil
}

func calculateCronSchedule(intervalInSeconds int) (string, error) {

	if intervalInSeconds <= 0 {
		return "", fmt.Errorf("intervalInSeconds must be greater than 0")
	}
	return fmt.Sprintf("*/%d * * * *", intervalInSeconds), nil
}
