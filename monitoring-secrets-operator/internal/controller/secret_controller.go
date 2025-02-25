package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	// Custom metric for counting reconciles
	reconcilesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "reconciles_total",
			Help: "Total number of reconciles by the controller",
		},
		[]string{"result"}, // Label to distinguish success/failure
	)
)

func init() {
	// Register the custom metrics with Prometheus
	prometheus.MustRegister(reconcilesTotal)
}

// SecretReconciler watches for Secret changes
type SecretReconciler struct {
	client.Client
}

// Reconcile is triggered when a Secret is created, updated, or deleted
func (r *SecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the existing Secret (old version)
	var oldSecret corev1.Secret
	err := r.Get(ctx, req.NamespacedName, &oldSecret)
	secretExists := err == nil

	// Fetch the new version of the Secret
	var newSecret corev1.Secret
	if err := r.Get(ctx, req.NamespacedName, &newSecret); err != nil {
		logger.Error(err, "❌ Unable to fetch Secret")
		// Increment reconcile failure counter
		reconcilesTotal.WithLabelValues("failure").Inc()
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Log Secret detected
	logger.Info("🔑 Secret detected", "Secret", newSecret.Name, "Namespace", newSecret.Namespace)

	// Compare old and new Secret data to check if the password has changed
	if secretExists {
		for key, newValue := range newSecret.Data {
			oldValue, exists := oldSecret.Data[key]

			// Convert values to hashes to avoid logging raw secrets
			newHash := hashString(newValue)
			oldHash := hashString(oldValue)

			if !exists || newHash != oldHash {
				logger.Info("🔑 Secret value changed", "Key", key, "Secret", newSecret.Name)
			}
		}
	}

	// Find Deployments that mount this Secret
	var deployments appsv1.DeploymentList
	if err := r.List(ctx, &deployments, client.InNamespace(newSecret.Namespace)); err != nil {
		logger.Error(err, "❌ Failed to list Deployments")
		// Increment reconcile failure counter
		reconcilesTotal.WithLabelValues("failure").Inc()
		return ctrl.Result{}, err
	}

	// Loop through each Deployment
	for _, deployment := range deployments.Items {
		usesSecret := false
		// Check if the Secret is being used in volumes of the Deployment
		for _, volume := range deployment.Spec.Template.Spec.Volumes {
			if volume.Secret != nil && volume.Secret.SecretName == newSecret.Name {
				usesSecret = true
				break
			}
		}

		if usesSecret {
			// Log Deployment using Secret
			logger.Info("🔄 Deployment uses updated Secret", "Deployment", deployment.Name)

			// Find Pods for the Deployment
			var pods corev1.PodList
			labelSelector := client.MatchingLabels(deployment.Spec.Template.Labels)
			if err := r.List(ctx, &pods, client.InNamespace(newSecret.Namespace), labelSelector); err != nil {
				logger.Error(err, fmt.Sprintf("❌ Failed to list Pods for Deployment %s", deployment.Name))
				// Increment reconcile failure counter
				reconcilesTotal.WithLabelValues("failure").Inc()
				return ctrl.Result{}, err
			}

			// Log details of the running pods
			logger.Info(fmt.Sprintf("🔄 Found %d Pods for Deployment %s", len(pods.Items), deployment.Name))

			// Check the readiness of the pods
			for _, pod := range pods.Items {
				if pod.Status.Phase == corev1.PodRunning {
					allContainersReady := true
					for _, container := range pod.Status.ContainerStatuses {
						if !container.Ready {
							allContainersReady = false
							break
						}
					}

					if allContainersReady {
						logger.Info("✅ Pod is ready", "Pod", pod.Name, "Deployment", deployment.Name)
					} else {
						logger.Info("⏳ Pod is not fully ready", "Pod", pod.Name, "Deployment", deployment.Name)
					}
				} else {
					logger.Info("⚠️ Pod is not in Running phase", "Pod", pod.Name, "Status", pod.Status.Phase)
				}
			}
		}
	}

	// Increment reconcile success counter
	reconcilesTotal.WithLabelValues("success").Inc()

	return ctrl.Result{}, nil
}

// hashString generates a SHA-256 hash of the given byte array
func hashString(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// SetupWithManager registers the controller with the manager
func (r *SecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// Register the metrics endpoint for Prometheus scraping
	go func() {
		http.Handle("/metrics", promhttp.Handler()) // Expose metrics at "/metrics"
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Log.Error(err, "Error starting metrics server")
		}
	}()

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Secret{}). // Watch Secret events
		Complete(r)
}
