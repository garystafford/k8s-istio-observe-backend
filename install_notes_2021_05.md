# Install Notes

Prerequisites

```shell
# https://eksctl.io
brew tap weaveworks/tap
brew upgrade weaveworks/tap/eksctl
eksctl version
# > 0.51.0
brew install yq
```

Installation

```shell
aws cloudformation create-stack \
    --stack-name istio-observability-demo \
    --template-body file://cloudformation/eks-vpc.yml \
    --capabilities CAPABILITY_NAMED_IAM

aws cloudformation describe-stacks \
    --stack-name istio-observability-demo \
    | jq -r '.Stacks[].Outputs'

# manually update vpc info in cluster.yaml file based on above cfn outputs

export AWS_ACCOUNT=$(aws sts get-caller-identity --output text --query 'Account')
export EKS_REGION="us-east-1"
export CLUSTER_NAME="eks-istio-observability-demo"

# yq e '.metadata.name = env(CLUSTER_NAME)' -i ./resources/other/cluster.yaml
# yq e '.metadata.region = env(EKS_REGION)' -i ./resources/other/cluster.yaml

eksctl create cluster -f ./resources/other/cluster.yaml
# 2021-05-26 12:37:12
# 2021-05-26 12:57:18
# 20 minutes

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
curl -o resources/aws/aws-load-balancer-controller-v220-all.yaml \
  https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.2.0/docs/install/v2_2_0_full.yaml

# https://docs.aws.amazon.com/eks/latest/userguide/aws-load-balancer-controller.html
# *** Modify the file aws-load-balancer-controller-v220-all.yaml per documentation #5 (two steps) ***
# - --cluster-name=eks-dev-cluster-istio

kubectl apply -f resources/aws/aws-load-balancer-controller-v220-all.yaml

kubectl get deployment -n kube-system aws-load-balancer-controller
# NAME                           READY   UP-TO-DATE   AVAILABLE   AGE
# aws-load-balancer-controller   1/1     1            1           55s
```

Create/Deploy ALB Policies/Roles

```shell
# *** manually update oidc-provider in two places in file ***

aws iam create-role \
  --role-name eks-alb-ingress-controller-eks-istio-observability-demo \
  --assume-role-policy-document file://resources/aws/trust-eks-istio-observability-demo.json

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observability-demo \
  --policy-arn="arn:aws:iam::${AWS_ACCOUNT}:policy/AWSLoadBalancerControllerIAMPolicy220"

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observability-demo \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy

aws iam attach-role-policy \
  --role-name eks-alb-ingress-controller-eks-istio-observability-demo \
  --policy-arn arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy
```

Deploy Istio

```
istioctl install --set profile=demo -y

kubectl apply -f resources/other/gateway.yaml -n dev
kubectl apply -f resources/other/destination-rules.yaml -n dev

istioctl analyze -n dev

kubectl apply -f $ISTIO_HOME/samples/addons
kubectl rollout status deployment/grafana -n istio-system

```

Deploy App

```shell
# namespaces
kubectl apply -f ./resources/other/namespaces.yaml
kubectl label namespace dev istio-injection=enabled
kubectl label namespace test istio-injection=enabled


# secrets
echo -n '{{mongodb.conn}}' | base64
echo -n '{{rabbitmq.conn}}' | base64

# *** Add to go-srv-demo.yaml file
kubectl apply -f ./resources/secrets/go-srv-demo-internal.yaml -n dev

kubectl apply -f ./resources/services/angular-ui.yaml -n dev

for service in a b c d e f g h; do
  kubectl apply -n dev -f "./resources/services/service-$service.yaml"
done


# ingress deploys the ALB
export ALB_CERT=$(aws acm list-certificates --certificate-statuses ISSUED \
  | jq -r '.CertificateSummaryList[] | select(.DomainName=="*.example-api.com") | .CertificateArn')
yq e '.metadata.annotations."alb.ingress.kubernetes.io/certificate-arn" = env(ALB_CERT)' -i resources/other/greetings-app-ingress.yaml

kubectl apply -f resources/other/greetings-app-ingress.yaml

```

