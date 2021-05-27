# EKS VPC Network Creation

```shell
aws cloudformation validate-template \
    --template-body file://cloudformation/eks-vpc.yml

aws cloudformation create-stack \
    --stack-name istio-observability-demo \
    --template-body file://cloudformation/eks-vpc.yml \
    --capabilities CAPABILITY_NAMED_IAM

aws cloudformation update-stack \
    --stack-name istio-observability-demo \
    --template-body file://cloudformation/eks-vpc.yml \
    --capabilities CAPABILITY_NAMED_IAM

aws cloudformation describe-stacks \
    --stack-name istio-observability-demo \
    | jq -r '.Stacks[].Outputs'
```
