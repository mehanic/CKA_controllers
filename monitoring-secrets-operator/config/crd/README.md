Explanation of customizeconfig.yaml
This file is a Kustomize configuration file that instructs Kustomize on how to substitute names and namespaces within Custom Resource Definitions (CRDs). It is particularly useful when deploying CRDs that reference Kubernetes Services via webhooks.

Key Sections
nameReference

yaml
Copy
Edit
nameReference:
  - kind: Service
    version: v1
    fieldSpecs:
      - kind: CustomResourceDefinition
        version: v1
        group: apiextensions.k8s.io
        path: spec/conversion/webhook/clientConfig/service/name
This tells Kustomize that if a Service name is referenced in a CRD webhook configuration, it should be updated when renaming services.
Use case: If your CRD webhook configuration references a service and the service name changes, Kustomize will automatically update the CRD.
namespace

yaml
Copy
Edit
namespace:
  - kind: CustomResourceDefinition
    version: v1
    group: apiextensions.k8s.io
    path: spec/conversion/webhook/clientConfig/service/namespace
    create: false
Ensures that the namespace of a webhook service in a CRD is updated when Kustomize modifies namespaces.
The create: false setting prevents Kustomize from creating this field if it doesn’t exist.
varReference

yaml
Copy
Edit
varReference:
  - path: metadata/annotations
Allows variable substitution within metadata annotations.
Useful for injecting dynamic values into annotations.
How It Improves Kustomization
✅ Automatic Name Updates: If a Service name changes, the corresponding webhook service reference in a CRD is updated automatically.
✅ Namespace Consistency: If you deploy to a different namespace, the webhook reference inside the CRD will match.
✅ Enhanced Customization: Variable references ensure metadata updates (useful for tracking deployment versions, labels, etc.).
✅ Avoids Manual Fixes: When deploying a Kubernetes operator or webhook, this prevents broken references in the CRD.

