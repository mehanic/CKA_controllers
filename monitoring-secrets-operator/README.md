### The primary purpose of this controller (SecretReconciler) is to manage and monitor changes to Secret resources in a Kubernetes cluster, particularly when the secret data (like passwords or tokens) is updated. 
The controller is designed to reconcile Secret changes by:

Watching for updates to Secret resources.

Hashing and comparing secret data to detect changes.

Identifying deployments using the secret and checking the readiness of the associated pods.

Exposing Prometheus metrics for tracking the controller's reconcile operations.

This provides an automated way to monitor and respond to changes in sensitive data (such as passwords) in Kubernetes while ensuring the associated workloads (like deployments and pods) are kept in sync with the updates.

Key Functionality of the Controller - Watch for Secret Changes: The controller watches for changes to Secret resources in the cluster, specifically detecting when a Secret is created, updated, or deleted.

When a Secret is updated, the Reconcile method is triggered. It checks if the secret has been modified and logs any changes. It compares the old and new versions of the Secret data to detect any differences in the data (e.g., a password or token change).

Instead of logging raw secret values, which could be sensitive, the controller hashes the secret values (using SHA-256) and compares the hashes to check for changes. This ensures sensitive data (such as passwords) is never exposed in logs.

The controller checks if the updated secret is being used by any Deployment in the same namespace. It does this by inspecting the Volumes in the Deployment spec to see if the secret is mounted as a volume.

Once it finds that a Deployment uses the updated secret, the controller proceeds to identify the associated Pods for that Deployment. It then checks the readiness of those pods to ensure that they are running and all containers are ready. If the pods are ready, it logs a success message; if not, it logs the status and checks their readiness state.

The controller exposes custom Prometheus metrics. It increments a reconciles_total counter every time a reconcile occurs, with the status labels indicating whether the reconcile was successful or failed. This allows Prometheus to track how many reconciles were performed and monitor their outcomes.

The controller serves a /metrics endpoint on port 8080 to expose the Prometheus metrics. This allows Prometheus to scrape the metrics and track the performance of the controller, particularly in terms of reconcile success and failure rates.

If the controller encounters any issues (e.g., failure to fetch a Secret, failure to list Deployments, or failure to list Pods), it logs the error and increments the failure counter in the Prometheus metrics.

The controller detects the update, compares the old and new versions of the secret, and if any changes are detected (based on the hash comparison), it checks if the secret is used in any deployments.

For each Deployment using the secret, the controller checks the associated Pods and their readiness. If the pods are ready, it logs that they are ready; otherwise, it logs their readiness state. The controller tracks its reconcile operations, recording both success and failure outcomes. These metrics are exposed to Prometheus for monitoring. A counter metric that tracks the total number of reconciles performed by the controller, with labels for success (success) or failure (failure).The metrics are exposed on the /metrics endpoint on port 8080, making them available for Prometheus to scrape.







# Steps to Create a Kubebuilder Project for Secret Monitoring
1. Install Prerequisites
```
curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/linux/amd64
chmod +x kubebuilder && sudo mv kubebuilder /usr/local/bin/
```
# Initialize the Kubebuilder Project
Create a new Kubebuilder project named "monitoring-secrets-operator":

```
mkdir monitoring-secrets-operator && cd monitoring-secrets-operator
kubebuilder init --domain mycompany.com --repo mycompany.com/monitoring-secrets-operator

#kubebuilder init --plugins go/v4 --domain example.org --owner "secretsync"

```

main.go → Entry point of the operator
config/ → Kubernetes manifests
controllers/ → Business logic (custom controller)
api/ → Custom Resource Definition (CRD)

# Create a Custom Resource Definition (CRD)
You need a CRD to define how you will configure secret monitoring.

```
kubebuilder create api --group security --version v1alpha1 --kind SecretMonitor
kubebuilder create api --group=core --version=v1 --kind=Secret --namespaced=false
```

api/v1alpha1/secretmonitor_types.go (CRD Schema)
controllers/secretmonitor_controller.go (Controller logic)

# Kubernetes & Podman Setup Guide

This guide provides step-by-step instructions on how to build, push, and deploy a containerized application using **Podman** and **Kubernetes**, along with managing **secrets**.

---

# Setting Up a Local Container Registry
Start a Podman container registry:
```sh
podman run -d -p 5000:5000 --restart=always --name registry registry:2

curl -X GET http://localhost:5000/v2/_catalog
```

# Building & Tagging the Image

```
podman build -t myrepo/password-app:1 . # main.go and Dockerfile in directory application

podman build -t my-image .
podman images

podman run -p 8080:8080 localhost/my-image:latest    #for test without -d

podman run -d -p 8080:8080 --name password-app myrepo/password-app:1

podman tag localhost/my-image:latest localhost:5000/my-image:v1

podman push localhost:5000/my-image:v1

podman push localhost:5000/my-image:v2
podman build -t localhost:5000/my-image:v2 .

podman build -t my-image .                                       
STEP 1/9: FROM golang:1.21
STEP 2/9: WORKDIR /app
--> f59eccfe31bf
STEP 3/9: COPY go.mod ./
--> d74de11c1fe9
STEP 4/9: RUN go mod download
--> b02e7ea9b69c
STEP 5/9: RUN go get github.com/prometheus/client_golang/prometheus github.com/prometheus/client_golang/prometheus/promhttp
--> 965aeb34c357
STEP 6/9: COPY *.go ./
--> 20f6254a89a9
STEP 7/9: RUN go build -o /go-podman-demo
--> f9d9f60b802f
STEP 8/9: EXPOSE 8080
--> a81de5713881
STEP 9/9: CMD [ "/go-podman-demo" ]
COMMIT my-image
--> 58aa1bf8577f
Successfully tagged localhost/my-image:latest
58aa1bf8577f614a3918a545f17fa522a12ffd79e26c2e585a00999ad3a910f3

╰─λ podman images                                                    
REPOSITORY                  TAG         IMAGE ID      CREATED        SIZE
localhost/my-image          latest      58aa1bf8577f  3 seconds ago  1.08 GB
docker.io/library/alpine    latest      aded1e1a5b37  11 days ago    8.13 MB
docker.io/library/golang    1.21        246ea1ed9cdb  6 months ago   838 MB
docker.io/library/registry  2           26b2eb03618e  17 months ago  26 MB

╰─λ podman tag localhost/my-image:latest localhost:5000/my-image:v3  


╰─λ podman images                                                     
REPOSITORY                  TAG         IMAGE ID      CREATED         SIZE
localhost:5000/my-image     v3          58aa1bf8577f  30 seconds ago  1.08 GB
localhost/my-image          latest      58aa1bf8577f  30 seconds ago  1.08 GB
docker.io/library/alpine    latest      aded1e1a5b37  11 days ago     8.13 MB
docker.io/library/golang    1.21        246ea1ed9cdb  6 months ago    838 MB
docker.io/library/registry  2           26b2eb03618e  17 months ago   26 MB
╰─λ                                                                   

podman push localhost:5000/my-image:v3                            
Getting image source signatures
Copying blob baeb38530ae9 skipped: already exists  
Copying blob a2d96588545f skipped: already exists  
Copying blob fc28d69664c6 done   | 
Copying blob bd9ddc54bea9 skipped: already exists  
Copying blob e2deec760404 done   | 
Copying blob 2794323d954c skipped: already exists  
Copying blob 5c19f166f90b done   | 
Copying blob 35494cb8a948 skipped: already exists  
Copying blob 44e7b3030cd8 done   | 
Copying blob 8ed96c8f5b6c done   | 
Copying blob 0e47201eaf58 skipped: already exists  
Copying blob 33948621fc9e skipped: already exists  
Copying config 58aa1bf857 done   | 
Writing manifest to image destination


```
# Configuring Kubernetes Deployment

```
kubectl apply -f deployment.yaml  # yaml file in directory kubernetes-files

kubectl get all

kubectl get pods -l app=my-app

kubectl logs -f $(kubectl get pods -l app=my-app -o jsonpath='{.items[0].metadata.name}')

sudo make run # in root directory

ls                                                               
api/          bin/  config/     go.mod  hack/      kubernetes-files/  PROJECT    test/
application/  cmd/  Dockerfile  go.sum  internal/  Makefile           README.md
```

# Main logic is wiriting with golang in  internal/controller/secret_controller.go and then added some paramethers which responsible for some features in cmd/main.go 

# After execute in terminal we see this info
```
2025-02-24T17:17:46+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"cilium-ca","namespace":"kube-system"}, "namespace": "kube-system", "name": "cilium-ca", "reconcileID": "b5d21669-5dfb-4e22-bf80-6b159cdd7fe9", "Secret": "cilium-ca", "Namespace": "kube-system"}
2025-02-24T17:17:46+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"sh.helm.release.v1.cilium.v4","namespace":"kube-system"}, "namespace": "kube-system", "name": "sh.helm.release.v1.cilium.v4", "reconcileID": "9c4ea236-541c-4e5a-8142-5d4c4e334ffc", "Secret": "sh.helm.release.v1.cilium.v4", "Namespace": "kube-system"}
2025-02-24T17:17:46+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"alertmanager-prometheus-operator-kube-p-alertmanager-tls-assets-0","namespace":"prometheus"}, "namespace": "prometheus", "name": "alertmanager-prometheus-operator-kube-p-alertmanager-tls-assets-0", "reconcileID": "ab5e51c5-d0b2-4cc9-8133-ef06ab3e3578", "Secret": "alertmanager-prometheus-operator-kube-p-alertmanager-tls-assets-0", "Namespace": "prometheus"}
2025-02-24T17:17:46+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-operator-kube-p-admission", "reconcileID": "87930ec4-57ee-4ae1-a702-5815c156032a", "Secret": "prometheus-operator-kube-p-admission", "Namespace": "prometheus"}
2025-02-24T17:17:46+01:00	INFO	🔄 Deployment uses updated Secret	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-operator-kube-p-admission", "reconcileID": "87930ec4-57ee-4ae1-a702-5815c156032a", "Deployment": "prometheus-operator-kube-p-operator"}
2025-02-24T17:17:46+01:00	INFO	🔄 Found 1 Pods for Deployment prometheus-operator-kube-p-operator	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-operator-kube-p-admission", "reconcileID": "87930ec4-57ee-4ae1-a702-5815c156032a"}
2025-02-24T17:17:46+01:00	INFO	✅ Pod is ready	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-operator-kube-p-admission", "reconcileID": "87930ec4-57ee-4ae1-a702-5815c156032a", "Pod": "prometheus-operator-kube-p-operator-78b875fd67-jfc64", "Deployment": "prometheus-operator-kube-p-operator"}
2025-02-24T17:17:46+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-prometheus-operator-kube-p-prometheus-web-config","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-prometheus-operator-kube-p-prometheus-web-config", "reconcileID": "1a90aae4-01b5-4d0c-9bee-f7092acdcd7a", "Secret": "prometheus-prometheus-operator-kube-p-prometheus-web-config", "Namespace": "prometheus"}


2025-02-25T19:38:05+01:00	INFO	✅ Pod is ready	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"prometheus"}, "namespace": "prometheus", "name": "prometheus-operator-kube-p-admission", "reconcileID": "2827694a-050e-4b78-ae0c-a5e37e966d79", "Pod": "prometheus-operator-kube-p-operator-78b875fd67-jfc64", "Deployment": "prometheus-operator-kube-p-operator"}
2025-02-25T19:38:05+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"my-app-secret","namespace":"default"}, "namespace": "default", "name": "my-app-secret", "reconcileID": "5789e9b5-61be-45ca-9b05-9f4cea7aa32c", "Secret": "my-app-secret", "Namespace": "default"}
2025-02-25T19:38:05+01:00	INFO	🔄 Deployment uses updated Secret	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"my-app-secret","namespace":"default"}, "namespace": "default", "name": "my-app-secret", "reconcileID": "5789e9b5-61be-45ca-9b05-9f4cea7aa32c", "Deployment": "my-app-deployment"}
2025-02-25T19:38:05+01:00	INFO	🔄 Found 1 Pods for Deployment my-app-deployment	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"my-app-secret","namespace":"default"}, "namespace": "default", "name": "my-app-secret", "reconcileID": "5789e9b5-61be-45ca-9b05-9f4cea7aa32c"}
2025-02-25T19:38:05+01:00	INFO	✅ Pod is ready	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"my-app-secret","namespace":"default"}, "namespace": "default", "name": "my-app-secret", "reconcileID": "5789e9b5-61be-45ca-9b05-9f4cea7aa32c", "Pod": "my-app-deployment-85bcc75bbc-mzbrb", "Deployment": "my-app-deployment"}
2025-02-25T19:38:05+01:00	INFO	🔑 Secret detected	{"controller": "secret", "controllerGroup": "", "controllerKind": "Secret", "Secret": {"name":"prometheus-operator-kube-p-admission","namespace":"default"}, "namespace": "default", "name": "prometheus-operator-kube-p-admission", "reconcileID": "9d5470fd-7895-4abe-95ae-b9186cb967b9", "Secret": "prometheus-operator-kube-p-admission", "Namespace": "default"}


```
# Working with Kubernetes Secrets

```
echo -n 'your-password' | base64
kubectl create secret generic my-app-secret --from-literal=password='your-password'

kubectl get secret my-app-secret -o jsonpath='{.data.password}' | base64 --decode

kubectl patch secret my-app-secret -n default --type='json' -p='[{"op": "replace", "path": "/data/password", "value": "eW91ci1wYXNzd29yZA=="}]'

echo -n 'your-capital' | base64


kubectl get secret my-app-secret -o jsonpath='{.data.password}' | base64 --decode

kubectl exec -it $(kubectl get pods -l app=my-app -o jsonpath='{.items[0].metadata.name}') -- cat /etc/secret-volume/password

```
# Managing Deployments

```
kubectl get deployment my-app-deployment

kubectl scale deployment my-app-deployment --replicas=3

kubectl rollout restart deployment my-app-deployment

kubectl get pods --selector=app=my-app -o wide

```
# Debugging Tips

```
kubectl exec -it $(kubectl get pods -l app=my-app -o jsonpath='{.items[0].metadata.name}') -- ls -l /etc/secret-volume

kubectl logs -f $(kubectl get pods -l app=my-app -o jsonpath='{.items[0].metadata.name}')

kubectl describe pod $(kubectl get pods -l app=my-app -o jsonpath='{.items[0].metadata.name}')
```
---


###  Add the Prometheus Helm Repository

```sh
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
kubectl create namespace prometheus
helm install prometheus-operator prometheus-community/kube-prometheus-stack -n prometheus
kubectl port-forward -n prometheus svc/prometheus-operator-grafana 3000:80
kubectl get secret -n prometheus prometheus-operator-grafana -o jsonpath="{.data.admin-password}" | base64 --decode
Username: admin

kubectl port-forward -n prometheus svc/prometheus-operator-kube-p-prometheus 9090:9090

kubectl port-forward -n prometheus svc/prometheus-operator-kube-p-alertmanager 9093:9093

kubectl port-forward -n prometheus svc/prometheus-operator-prometheus-node-exporter 9100:9100
```
# Running Multiple Port Forwards at Once

```
kubectl port-forward -n prometheus svc/prometheus-operator-grafana 3000:80 &
kubectl port-forward -n prometheus svc/prometheus-operator-kube-p-prometheus 9090:9090 &
kubectl port-forward -n prometheus svc/prometheus-operator-kube-p-alertmanager 9093:9093 &
```
netstat -tulnp | grep LISTEN

----




# Custom Resource Definition (CRD) Management for Secrets Operator

This guide explains how to verify, delete, reapply, and use a **Custom Resource Definition (CRD)** for managing Kubernetes secrets.



## Step 1: Verify Existing CRD


Before applying a new CRD, check if the old CRD is already installed:
```sh
kubectl get crds  secrets.core.mycompany.com

sudo kubectl get crds | grep managedsecrets.core.mycompany.com

#If it exists, delete it to prevent conflicts.

kubectl delete crd secrets.core.mycompany.com

```
# Apply the New CRD

```
kubectl apply -f your_crd_file.yaml


kubectl get crds 
secrets.core.mycompany.com       YYYY-MM-DDTHH:MM:SSZ

managedsecrets.core.mycompany.com            2025-02-24T15:31:03Z

```
# Creating and Managing Custom Resources
Once the CRD is applied, you can create custom resources using YAML. Create a Custom Secret Resource
Example my-secret.yaml:
```
yaml
Copy
Edit
apiVersion: core.mycompany.com/v1
kind: ManagedSecret
metadata:
  name: my-custom-secret
spec:
  secretName: my-app-secret
  data:
    username: admin
    password: supersecurepassword
  labels:
    environment: production

```
Apply it using:

kubectl apply -f my-secret.yaml

# Verify Custom Resources
To check if your ManagedSecrets are correctly created:


kubectl get managedsecrets or using the short name:
```
kubectl get msec

NAME               AGE
my-custom-secret   2m
To describe a specific secret:

kubectl describe managedsecret my-custom-secret
kubectl delete managedsecret my-custom-secret

sudo kubectl apply -f core.mycompany.com_secrets.yaml
sudo kubectl delete -f core.mycompany.com_secrets.yaml
```


# ServiceMonitor resource
The error indicates that the ServiceMonitor resource is not recognized in your cluster, which means the Prometheus Operator CRDs are not installed. ServiceMonitor is a custom resource provided by the Prometheus Operator, and your cluster must have the Prometheus Operator CRDs installed before you apply monitor.yaml.

Check if the ServiceMonitor CRD is Installed
```
kubectl get crds | grep servicemonitors
```

If the CRD is missing, install the Prometheus Operator, which provides the ServiceMonitor resource. You can do this in multiple ways:

# Step: Retry applying monitor.yaml
Once the CRD is available, you can now apply your monitor.yaml file:

```
kubectl apply -f monitor.yaml
```
If everything is correctly set up, it should successfully create the ServiceMonitor resource.

# Step: Verify the ServiceMonitor
Check if the ServiceMonitor was created successfully:

```
kubectl get servicemonitor -n system
```
Ensure Prometheus is correctly scraping the /metrics endpoint.
Check Prometheus logs if the service is being monitored.
You can check logs using:

```
kubectl logs -l app=prometheus -n monitoring

kubectl patch servicemonitor controller-manager-metrics-monitor -n default --type='merge' -p '{"metadata":{"labels":{"release":"prometheus-operator"}}}'

 sudo kubectl apply -f monitor.yaml
servicemonitor.monitoring.coreos.com/controller-manager-metrics-monitor created

  sudo kubectl get servicemonitor -A | grep controller-manager-metrics-monitor
default      controller-manager-metrics-monitor                   22s

 kubectl get servicemonitors -n default --show-labels
NAME                                 AGE   LABELS
controller-manager-metrics-monitor   85s   app.kubernetes.io/managed-by=kustomize,app.kubernetes.io/name=monitoring-secrets-operator,control-plane=controller-manager,release=prometheus-operator

```

# Step 1: Check ServiceMonitor Discovery
Open Prometheus UI:
```
kubectl port-forward -n prometheus svc/prometheus-operated 9090
```
Now, go to http://localhost:9090 in your browser.

Navigate to Status → Service Discovery (http://localhost:9090/service-discovery)


Look for servicemonitor/default/controller-manager-metrics-monitor
If it's listed, Prometheus has discovered it!
Step 2: Check Target Status
Go to Status → Targets (http://localhost:9090/targets)
Find controller-manager-metrics-monitor
🟢 UP: If the target is UP, Prometheus is scraping it successfully.
🔴 DOWN: If it’s DOWN, hover over the ❌ error to see the issue (e.g., connection refused, TLS errors, etc.).

# Step 3: Run Queries to Verify Data
Go to Graph (http://localhost:9090/graph).
Enter a metric name related to controller-manager-metrics-monitor, such as:
```
up{job="controller-manager-metrics-monitor"}
```
Click Execute to see if metrics are being collected.
Step 4: Debug if Not Working
If you don’t see it in Targets, check logs:

Check Prometheus Logs

```
kubectl logs -n prometheus statefulset/prometheus-operated
```
Look for errors related to controller-manager-metrics-monitor.

Check ServiceMonitor Matching

```
kubectl get servicemonitor -n default controller-manager-metrics-monitor -o yaml
```
Ensure selector.matchLabels matches the labels of the corresponding Service.

```
# Check If Metrics Are Available

kubectl get svc -n default
# Find the service exposing /metrics
Try manually curling it:

kubectl run -it --rm --image=curlimages/curl debug -- sh
curl -k https://controller-manager-service.default.svc:20000/metrics


kubectl run -it --rm --image=curlimages/curl debug -- curl -k https://controller-manager-service.default.svc:20000/metrics

```
--------------

#  You can verify that the /metrics endpoint is working by visiting:

```
─(~/monitoring-secrets-operator/kubernetes-files)
 └> $ curl http://192.168.0.227:30080/metrics
# HELP go_gc_duration_seconds A summary of the wall-time pause (stop-the-world) duration in garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 0
go_gc_duration_seconds{quantile="0.25"} 0
go_gc_duration_seconds{quantile="0.5"} 0
go_gc_duration_seconds{quantile="0.75"} 0
go_gc_duration_seconds{quantile="1"} 0
go_gc_duration_seconds_sum 0
go_gc_duration_seconds_count 0
# HELP go_gc_gogc_percent Heap size target percentage configured by the user, otherwise 100. This value is set by the GOGC environment variable, and the runtime/debug.SetGCPercent function. Sourced from /gc/gogc:percent.
# TYPE go_gc_gogc_percent gauge
go_gc_gogc_percent 100
# HELP go_gc_gomemlimit_bytes Go runtime memory limit configured by the user, otherwise math.MaxInt64. This value is set by the GOMEMLIMIT environment variable, and the runtime/debug.SetMemoryLimit function. Sourced from /gc/gomemlimit:bytes.
# TYPE go_gc_gomemlimit_bytes gauge
go_gc_gomemlimit_bytes 9.223372036854776e+18
# HELP go_goroutines Number of goroutines that currently exist.
# TYPE go_goroutines gauge
go_goroutines 7
# HELP go_info Information about the Go environment.
# TYPE go_info gauge
go_info{version="go1.21.13"} 1
# HELP go_memstats_alloc_bytes Number of bytes allocated in heap and currently in use. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_alloc_bytes gauge
go_memstats_alloc_bytes 226800
# HELP go_memstats_alloc_bytes_total Total number of bytes allocated in heap until now, even if released already. Equals to /gc/heap/allocs:bytes.
# TYPE go_memstats_alloc_bytes_total counter
go_memstats_alloc_bytes_total 226800
# HELP go_memstats_buck_hash_sys_bytes Number of bytes used by the profiling bucket hash table. Equals to /memory/classes/profiling/buckets:bytes.
# TYPE go_memstats_buck_hash_sys_bytes gauge
go_memstats_buck_hash_sys_bytes 4350
# HELP go_memstats_frees_total Total number of heap objects frees. Equals to /gc/heap/frees:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_frees_total counter
go_memstats_frees_total 0
# HELP go_memstats_gc_sys_bytes Number of bytes used for garbage collection system metadata. Equals to /memory/classes/metadata/other:bytes.
# TYPE go_memstats_gc_sys_bytes gauge
go_memstats_gc_sys_bytes 2.477616e+06
# HELP go_memstats_heap_alloc_bytes Number of heap bytes allocated and currently in use, same as go_memstats_alloc_bytes. Equals to /memory/classes/heap/objects:bytes.
# TYPE go_memstats_heap_alloc_bytes gauge
go_memstats_heap_alloc_bytes 226800
# HELP go_memstats_heap_idle_bytes Number of heap bytes waiting to be used. Equals to /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_idle_bytes gauge
go_memstats_heap_idle_bytes 2.023424e+06
# HELP go_memstats_heap_inuse_bytes Number of heap bytes that are in use. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes
# TYPE go_memstats_heap_inuse_bytes gauge
go_memstats_heap_inuse_bytes 1.744896e+06
# HELP go_memstats_heap_objects Number of currently allocated objects. Equals to /gc/heap/objects:objects.
# TYPE go_memstats_heap_objects gauge
go_memstats_heap_objects 495
# HELP go_memstats_heap_released_bytes Number of heap bytes released to OS. Equals to /memory/classes/heap/released:bytes.
# TYPE go_memstats_heap_released_bytes gauge
go_memstats_heap_released_bytes 2.023424e+06
# HELP go_memstats_heap_sys_bytes Number of heap bytes obtained from system. Equals to /memory/classes/heap/objects:bytes + /memory/classes/heap/unused:bytes + /memory/classes/heap/released:bytes + /memory/classes/heap/free:bytes.
# TYPE go_memstats_heap_sys_bytes gauge
go_memstats_heap_sys_bytes 3.76832e+06
# HELP go_memstats_last_gc_time_seconds Number of seconds since 1970 of last garbage collection.
# TYPE go_memstats_last_gc_time_seconds gauge
go_memstats_last_gc_time_seconds 0
# HELP go_memstats_mallocs_total Total number of heap objects allocated, both live and gc-ed. Semantically a counter version for go_memstats_heap_objects gauge. Equals to /gc/heap/allocs:objects + /gc/heap/tiny/allocs:objects.
# TYPE go_memstats_mallocs_total counter
go_memstats_mallocs_total 495
# HELP go_memstats_mcache_inuse_bytes Number of bytes in use by mcache structures. Equals to /memory/classes/metadata/mcache/inuse:bytes.
# TYPE go_memstats_mcache_inuse_bytes gauge
go_memstats_mcache_inuse_bytes 14400
# HELP go_memstats_mcache_sys_bytes Number of bytes used for mcache structures obtained from system. Equals to /memory/classes/metadata/mcache/inuse:bytes + /memory/classes/metadata/mcache/free:bytes.
# TYPE go_memstats_mcache_sys_bytes gauge
go_memstats_mcache_sys_bytes 15600
# HELP go_memstats_mspan_inuse_bytes Number of bytes in use by mspan structures. Equals to /memory/classes/metadata/mspan/inuse:bytes.
# TYPE go_memstats_mspan_inuse_bytes gauge
go_memstats_mspan_inuse_bytes 46032
# HELP go_memstats_mspan_sys_bytes Number of bytes used for mspan structures obtained from system. Equals to /memory/classes/metadata/mspan/inuse:bytes + /memory/classes/metadata/mspan/free:bytes.
# TYPE go_memstats_mspan_sys_bytes gauge
go_memstats_mspan_sys_bytes 48888
# HELP go_memstats_next_gc_bytes Number of heap bytes when next garbage collection will take place. Equals to /gc/heap/goal:bytes.
# TYPE go_memstats_next_gc_bytes gauge
go_memstats_next_gc_bytes 4.194304e+06
# HELP go_memstats_other_sys_bytes Number of bytes used for other system allocations. Equals to /memory/classes/other:bytes.
# TYPE go_memstats_other_sys_bytes gauge
go_memstats_other_sys_bytes 966906
# HELP go_memstats_stack_inuse_bytes Number of bytes obtained from system for stack allocator in non-CGO environments. Equals to /memory/classes/heap/stacks:bytes.
# TYPE go_memstats_stack_inuse_bytes gauge
go_memstats_stack_inuse_bytes 425984
# HELP go_memstats_stack_sys_bytes Number of bytes obtained from system for stack allocator. Equals to /memory/classes/heap/stacks:bytes + /memory/classes/os-stacks:bytes.
# TYPE go_memstats_stack_sys_bytes gauge
go_memstats_stack_sys_bytes 425984
# HELP go_memstats_sys_bytes Number of bytes obtained from system. Equals to /memory/classes/total:byte.
# TYPE go_memstats_sys_bytes gauge
go_memstats_sys_bytes 7.707664e+06
# HELP go_sched_gomaxprocs_threads The current runtime.GOMAXPROCS setting, or the number of operating system threads that can execute user-level Go code simultaneously. Sourced from /sched/gomaxprocs:threads.
# TYPE go_sched_gomaxprocs_threads gauge
go_sched_gomaxprocs_threads 12
# HELP go_threads Number of OS threads created.
# TYPE go_threads gauge
go_threads 8
# HELP password_access_total Total number of times the secret password is accessed
# TYPE password_access_total counter
password_access_total{status="success"} 8
# HELP process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE process_cpu_seconds_total counter
process_cpu_seconds_total 0.01
# HELP process_max_fds Maximum number of open file descriptors.
# TYPE process_max_fds gauge
process_max_fds 1.048576e+06
# HELP process_network_receive_bytes_total Number of bytes received by the process over the network.
# TYPE process_network_receive_bytes_total counter
process_network_receive_bytes_total 7595
# HELP process_network_transmit_bytes_total Number of bytes sent by the process over the network.
# TYPE process_network_transmit_bytes_total counter
process_network_transmit_bytes_total 4650
# HELP process_open_fds Number of open file descriptors.
# TYPE process_open_fds gauge
process_open_fds 8
# HELP process_resident_memory_bytes Resident memory size in bytes.
# TYPE process_resident_memory_bytes gauge
process_resident_memory_bytes 9.1136e+06
# HELP process_start_time_seconds Start time of the process since unix epoch in seconds.
# TYPE process_start_time_seconds gauge
process_start_time_seconds 1.74050250048e+09
# HELP process_virtual_memory_bytes Virtual memory size in bytes.
# TYPE process_virtual_memory_bytes gauge
process_virtual_memory_bytes 1.652068352e+09
# HELP process_virtual_memory_max_bytes Maximum amount of virtual memory available in bytes.
# TYPE process_virtual_memory_max_bytes gauge
process_virtual_memory_max_bytes 1.8446744073709552e+19
# HELP promhttp_metric_handler_requests_in_flight Current number of scrapes being served.
# TYPE promhttp_metric_handler_requests_in_flight gauge
promhttp_metric_handler_requests_in_flight 1
# HELP promhttp_metric_handler_requests_total Total number of scrapes by HTTP status code.
# TYPE promhttp_metric_handler_requests_total counter
promhttp_metric_handler_requests_total{code="200"} 0
promhttp_metric_handler_requests_total{code="500"} 0
promhttp_metric_handler_requests_total{code="503"} 0


 └> $ curl http://192.168.0.227:30080/metrics | grep password_access_total{status="success"}
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  8731    0  8731    0     0  5335k      0 --:--:-- --:--:-- --:--:-- 8526k

 └> $ curl http://192.168.0.227:30080/metrics | grep password_access_total
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  8740    0  8740    0    # HELP password_access_total Total number of times the secret password is accessed
 0# TYPE password_access_total counter
 password_access_total{status="success"} 8 it is number how much change password
 5095k      0 --:--:-- --:--:-- --:--:-- 8535k


 └> $ kubectl patch secret my-app-secret -n default --type='json' -p='[{"op": "replace", "path": "/data/password", "value": "eW91ci1wb3NzaWJpbGl0eQ=="}]'
secret/my-app-secret patched

 └> $ curl http://192.168.0.227:30080/metrics | grep password_access_total
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100  8741    0  8741    0     0  5111k      0 --:--:-- # HELP password_access_total Total number of times the secret password is accessed
--:# TYPE password_access_total counter
--password_access_total{status="success"} 12 # it is number how much change password
:-- --:--:-- 8536k



```

Ensure Webhook Service Exists

Your CRD's spec.conversion.webhook.clientConfig references a Service for the webhook.
Confirm you have a Service defined for your webhook (e.g., monitoring-secrets-webhook).

Example Service:

```
apiVersion: v1
kind: Service
metadata:
  name: monitoring-secrets-webhook
  namespace: monitoring
spec:
  selector:
    app: monitoring-secrets-webhook
  ports:
    - protocol: TCP
      port: 443
      targetPort: 8443
Modify your CRD (bases/core.mycompany.com_secrets.yaml) to include a webhook conversion config.
```
Example:

```
apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: secrets.core.mycompany.com
spec:
  group: core.mycompany.com
  names:
    kind: Secret
    plural: secrets
  scope: Namespaced
  conversion:
    strategy: Webhook
    webhook:
      clientConfig:
        service:
          name: monitoring-secrets-webhook  # Kustomize will substitute this
          namespace: monitoring             # Kustomize will substitute this
          path: /convert
        caBundle: Cg==
```

Your kustomization.yaml should declare the webhook CRD and the config file.

Example:

resources:
  - bases/core.mycompany.com_secrets.yaml  # Your CRD file

# patches:(not finish)

Uncomment if you're enabling webhooks
  - path: webhook_patch.yaml

configurations:
  - kustomizeconfig.yaml
Apply the Configuration
Once you've configured everything, apply it:

kubectl apply -k .
Then verify:





## In Kubebuilder, you should apply monitor.yaml after deploying the CRDs and before running the controller. Here’s the correct sequence:

1️⃣ Setup Your Controller and CRDs
Before applying the ServiceMonitor (monitor.yaml), you need to ensure that:

The Custom Resource Definitions (CRDs) are installed.
The controller is deployed and running, exposing metrics.
Step-by-Step Order:
Generate CRDs (if not already created):

make manifests
Apply CRDs to the cluster:


kubectl apply -f config/crd/bases/
Deploy the Controller:
If running locally:

make run
If deploying via Kubernetes:

make docker-build docker-push IMG=<your_image>
make deploy IMG=<your_image>
 Apply monitor.yaml
Now that the controller is running and exposing metrics, apply your ServiceMonitor:


kubectl apply -f monitor.yaml
3Verify That Metrics Are Being Scraped
Check if ServiceMonitor is applied correctly:

kubectl get servicemonitor -A
Check if Prometheus is discovering the controller metrics:

kubectl get servicemonitors -n <prometheus-namespace>
Check logs if metrics are not being scraped:

kubectl logs -n <namespace> <prometheus-pod-name>

Summary:

✔ Apply CRDs first

✔ Deploy Controller next

✔ Apply monitor.yaml 
after controller is running

✔ Verify metrics in Prometheus

