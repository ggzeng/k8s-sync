#! /bin/sh

function help() {
  echo "error: $1"
  echo ""
  echo "Usage: $0 <apply|delete> <image-tag> <ns>"
  exit 1
}

[ "$1" != "apply" ] && [ "$1" != "delete" ] && help "unkonw cmd type \"$1\""
CMD="$1"

[ "x$2" == "x" ] && help "need a image tag"
TAG="$2"

if [ "x$3" == "x" ] ; then
  NS="default"
else
  NS="$3"
fi
echo "use namespace $NS"

if [ "$CMD" == "apply" ] ; then
  if kubectl get configmap -n $NS | grep -q cloud-gateway-gwadmin ; then
    kubectl create configmap cloud-gateway-gwadmin --from-file=../configs/settings.prod.yaml -n $NS -o yaml --dry-run=client | kubectl replace -f -
  else
    kubectl create configmap cloud-gateway-gwadmin --from-file=../configs/settings.prod.yaml -n $NS
  fi
else
    kubectl delete configmap cloud-gateway-gwadmin -n $NS
fi

sed -e "s/\<image-tag\>/${TAG}/" gwadmin.yaml | sed -e "s/\<ns\>/${NS}/" > .deploy.yaml
kubectl "$CMD" -n $NS -f .deploy.yaml
if [ $? -eq 0 ] ; then
  echo ""
  echo "Success."
else
  echo ""
  echo "Failed!"
fi
