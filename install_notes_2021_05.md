# Install Notes

References

-<https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.2/>


Prerequisites

```shell
# https://eksctl.io
brew tap weaveworks/tap
brew upgrade weaveworks/tap/eksctl
eksctl version
# > 0.51.0

# brew install kubernetes-cli awscli yq jq
# brew upgrade kubernetes-cli awscli yq jq go
```

Installation

```shell
export AWS_ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
export EKS_REGION="us-east-1"
export CLUSTER_NAME="istio-observe-demo"

yq e '.metadata.name = env(CLUSTER_NAME)' -i ./resources/other/cluster.yaml
yq e '.metadata.region = env(EKS_REGION)' -i ./resources/other/cluster.yaml

eksctl create cluster -f ./resources/other/cluster.yaml
# 2021-05-27 21:21:13
# 2021-05-27 22:05:49

aws eks --region ${EKS_REGION} update-kubeconfig --name ${CLUSTER_NAME}

kubectl cluster-info

eksctl utils describe-stacks \
  --region ${EKS_REGION} --cluster ${CLUSTER_NAME}
```

ALB - Install alb-ingress-controller

```shell
# *** Not needed given IODC specified in cluster file"
#eksctl utils associate-iam-oidc-provider \
#    --region ${EKS_REGION} \
#    --cluster ${CLUSTER_NAME} \
#    --approve

# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
# 1x
curl -o resources/aws/iam-policy.json \
  https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.2.0/docs/install/iam_policy.json

# 1x
aws iam create-policy \
    --policy-name AWSLoadBalancerControllerIAMPolicy220 \
    --policy-document file://resources/aws/iam-policy.json

eksctl create iamserviceaccount \
  --region ${EKS_REGION} \
  --cluster ${CLUSTER_NAME} \
  --namespace=kube-system \
  --name=aws-load-balancer-controller \
  --attach-policy-arn=arn:aws:iam::${AWS_ACCOUNT}:policy/AWSLoadBalancerControllerIAMPolicy220 \
  --override-existing-serviceaccounts \
  --approve

kubectl apply --validate=false \
  -f https://github.com/jetstack/cert-manager/releases/download/v1.3.1/cert-manager.yaml

# 1x
curl -o resources/other/aws-load-balancer-controller-v220-all.yaml \
  https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.2.0/docs/install/v2_2_0_full.yaml

# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
# *** Modify the file aws-load-balancer-controller-v220-all.yaml per documentation #5 (two steps) ***
# - --cluster-name=eks-dev-cluster-istio

kubectl apply -f resources/other/aws-load-balancer-controller-v220-all.yaml

kubectl get deployment -n kube-system aws-load-balancer-controller
# NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
# aws-load-balancer-controller   1/1     1            1           55s
```

Create/Deploy ALB Policies/Roles

```shell
# *** manually update oidc-provider in two places in file ***
# use
# aws eks describe-cluster --name ${CLUSTER_NAME}
# aws iam list-open-id-connect-providers

aws iam create-role \
  --role-name eks-alb-ingress-controller-eks-istio-observe-demo \
  --assume-role-policy-document file://resources/aws/trust-eks-istio-observe-demo.json

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observe-demo \
  --policy-arn="arn:aws:iam::${AWS_ACCOUNT}:policy/AWSLoadBalancerControllerIAMPolicy220"

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observe-demo \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observe-demo \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
```

Deploy Istio and ALB
```shell
# namespaces
kubectl apply -f ./resources/other/namespaces.yaml
kubectl label namespace dev istio-injection=enabled
kubectl label namespace test istio-injection=enabled

istioctl install --set profile=demo -y

# https://itnext.io/istio-external-aws-application-loadbalancer-and-istio-ingress-gateway-fce3bfd3202f
# https://stackoverflow.com/a/66627104/580268
kubectl -n istio-system edit svc istio-ingressgateway #NodePort and get port #
kubectl -n istio-system describe svc istio-ingressgateway
kubectl -n istio-system get deploy istio-ingressgateway -o yaml

kubectl apply -f resources/istio/gateway.yaml -n dev
kubectl apply -f resources/istio/destination-rules.yaml -n dev

# *** manually add 3 healthcheck lines with port from above, per instructions also above
kubectl apply -f resources/other/istio-ingress.yaml

# key command to confirm health of ALB config!
kubectl describe ingress.networking.k8s.io --all-namespaces

kubectl -n istio-system get ingress

kubectl apply -f $ISTIO_HOME/samples/addons
kubectl rollout status deployment/grafana -n istio-system
```

Deploy fluent bit

```shell
# https://aws.amazon.com/blogs/containers/fluent-bit-integration-in-cloudwatch-container-insights-for-eks/
kubectl apply -f https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/cloudwatch-namespace.yaml

ClusterName=${CLUSTER_NAME}
RegionName=${EKS_REGION}
FluentBitHttpPort='2020'
FluentBitReadFromHead='Off'
[[ ${FluentBitReadFromHead} = 'On' ]] && FluentBitReadFromTail='Off'|| FluentBitReadFromTail='On'
[[ -z ${FluentBitHttpPort} ]] && FluentBitHttpServer='Off' || FluentBitHttpServer='On'
kubectl create configmap fluent-bit-cluster-info \
--from-literal=cluster.name=${ClusterName} \
--from-literal=http.server=${FluentBitHttpServer} \
--from-literal=http.port=${FluentBitHttpPort} \
--from-literal=read.head=${FluentBitReadFromHead} \
--from-literal=read.tail=${FluentBitReadFromTail} \
--from-literal=logs.region=${RegionName} -n amazon-cloudwatch

kubectl apply -f https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/fluent-bit/fluent-bit.yaml

kubectl get pods -n amazon-cloudwatch

DASHBOARD_NAME=istio_observe_demo
REGION_NAME=${EKS_REGION}
CLUSTER_NAME=${CLUSTER_NAME}

curl https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/service/cwagent-prometheus/sample_cloudwatch_dashboards/fluent-bit/cw_dashboard_fluent_bit.json \
| sed "s/{{YOUR_AWS_REGION}}/${REGION_NAME}/g" \
| sed "s/{{YOUR_CLUSTER_NAME}}/${CLUSTER_NAME}/g" \
| xargs -0 aws cloudwatch put-dashboard --dashboard-name ${DASHBOARD_NAME} --dashboard-body


curl https://raw.githubusercontent.com/aws-samples/amazon-cloudwatch-container-insights/latest/k8s-deployment-manifest-templates/deployment-mode/daemonset/container-insights-monitoring/fluentd/fluentd.yaml | kubectl delete -f -
kubectl delete configmap cluster-info -n amazon-cloudwatch
```

Deploy App

```shell
# secrets
echo -n '{{mongodb.conn}}' | base64
echo -n '{{rabbitmq.conn}}' | base64

# *** Add to go-srv-demo.yaml file
kubectl apply -f ./resources/secrets/go-srv-demo-internal.yaml -n dev

kubectl apply -f ./resources/services/angular-ui.yaml -n dev

for service in a b c d e f g h; do
  kubectl apply -f "./resources/services/service-$service.yaml" -n dev
done


# ingress deploys the ALB
export ALB_CERT=$(aws acm list-certificates --certificate-statuses ISSUED \
  | jq -r '.CertificateSummaryList[] | select(.DomainName=="*.example-api.com") | .CertificateArn')
yq e '.metadata.annotations."alb.ingress.kubernetes.io/certificate-arn" = env(ALB_CERT)' -i resources/other/greetings-app-ingress.yaml

istioctl analyze -n dev

```

Delete Resources

```shell
eksctl delete cluster --name $CLUSTER_NAME

aws cloudformation delete-stack --stack-name istio-observability-demo
```
