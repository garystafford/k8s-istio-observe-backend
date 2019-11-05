# EKS Commands

```bash
eksctl create cluster --help

eksctl create cluster \
    --name prod \
    --version 1.14 \
    --region us-west-2 \
    --nodegroup-name standard-workers \
    --node-type t3.medium \
    --nodes 2 \
    --nodes-min 1 \
    --nodes-max 3 \
    --node-ami auto

eksctl utils describe-stacks --region=us-west-2 --cluster=prod

https://docs.aws.amazon.com/eks/latest/userguide/getting-started-eksctl.html
curl --silent --location "https://github.com/weaveworks/eksctl/releases/download/latest_release/eksctl_$(uname -s)_amd64.tar.gz" | tar xz -C /tmp
sudo mv /tmp/eksctl /usr/local/bin
eksctl version

aws eks list-clusters --region=us-west-2
aws sts get-caller-identity
export AWS_REGION=us-west-2
eksctl get cluster



curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
kubectl version
kubectl config view


aws eks update-kubeconfig --name prod --profile admin --region us-west-2
aws eks list-clusters --region us-west-2
aws eks delete-cluster --name=prod --region=us-west-2

https://istio.io/docs/setup/install/helm/
brew install kubernetes-helm
kubectl create namespace istio-system
helm repo add istio.io https://storage.googleapis.com/istio-release/releases/1.3.4/charts/
helm template install/kubernetes/helm/istio-init --name istio-init --namespace istio-system | kubectl apply -f -
kubectl get crds | grep 'istio.io' | wc -l
helm template install/kubernetes/helm/istio --name istio --namespace istio-system | kubectl apply -f -


curl -L https://git.io/getLatestIstio | ISTIO_VERSION=1.3.4 sh -
export PATH="$PATH:/Users/garystaf/istio-1.3.4/bin"
export ISTIO_HOME=~/istio-1.3.4


```
