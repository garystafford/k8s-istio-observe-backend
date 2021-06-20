# Reference Application Platform Helm Chart

Helm chart that will install all Kubernetes resources to the `dev` namespace. Place your environment specific values in the `values.yaml` first before apply chart to your k8s cluster. Note that this chart includes `HorizontalPodAutoscaler` resources, not discussed in the blog posts associated this source code.

```shell
# dry run
helm install ref-app ./ref-app-chart --namespace dev --debug --dry-run

# apply chart resources
helm install ref-app ./ref-app-chart --namespace dev | kubectl apply -f
```

Resources included in Helm Chart:

```text
.
└── dev
    ├── hpa
    │   ├── hpa-angular-ui.yaml
    │   ├── hpa-service-a.yaml
    │   ├── hpa-service-b.yaml
    │   ├── hpa-service-c.yaml
    │   ├── hpa-service-d.yaml
    │   ├── hpa-service-e.yaml
    │   ├── hpa-service-f.yaml
    │   ├── hpa-service-g.yaml
    │   └── hpa-service-h.yaml
    ├── istio
    │   ├── destination-rules.yaml
    │   ├── external-mesh-amazon-mq.yaml
    │   ├── external-mesh-document-db.yaml
    │   ├── gateway.yaml
    │   └── virtualservices.yaml
    ├── secrets
    │   └── secrets.yaml
    └── services
        ├── angular-ui.yaml
        ├── service-a.yaml
        ├── service-b.yaml
        ├── service-c.yaml
        ├── service-d.yaml
        ├── service-e.yaml
        ├── service-f.yaml
        ├── service-g.yaml
        └── service-h.yaml
```