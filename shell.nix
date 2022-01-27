with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    pkgs.kubectl
    pkgs.kubernetes-helm
    pkgs.kompose
    pkgs.killall
    pkgs.coreutils
    pkgs.kube3d
  ];

  shellHook = ''
trap 'k3d cluster stop gm-cluster; k3d cluster delete gm-cluster;  killall kubectl;' EXIT
docker --version
if docker --version; then
  echo "Docker Daemon is running" 
else
  echo "Docker Daemon is not running. Please install and run it on your system." 
  exit 0;
fi

docker login docker.greymatter.io

k3d cluster create gm-cluster -a 1 -p 30000:10808@loadbalancer
k3d cluster start gm-cluster
kubectl get nodes

#Make sure we are using the right context
kubectl config use-context k3d-cluster

# Install k8s dashboard
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.4.0/aio/deploy/recommended.yaml
kubectl proxy &

echo "apiVersion: v1
kind: ServiceAccount
metadata:
  name: admin-user
  namespace: kubernetes-dashboard" | kubectl apply -f -

echo "apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: admin-user
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: admin-user
  namespace: kubernetes-dashboard" | kubectl apply -f - 

if [[ "$OSTYPE" == "darwin"* ]]; then
  # Mac OSX
  if [[ -z "$GREYMATTER_REGISTRY_USERNAME" ]]; then
    echo "Docker Username(email): "
    read GREYMATTER_REGISTRY_USERNAME
  fi
  if [[ -z "$GREYMATTER_REGISTRY_PASSWORD" ]]; then
    echo "Docker Password: "
    read -s GREYMATTER_REGISTRY_PASSWORD
  fi
  kubectl create secret docker-registry regcred \
    --docker-server=docker.greymatter.io \
    --docker-username=$GREYMATTER_REGISTRY_USERNAME \
    --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
    --docker-email=$GREYMATTER_REGISTRY_USERNAME \
    -n default
else
  kubectl create secret generic regcred \
   --from-file=.dockerconfigjson=$HOME/.docker/config.json \
   --type=kubernetes.io/dockerconfigjson \
   -n default
fi

echo "Just a moment..."
sleep 5;
echo "Adding private credentials to service account."
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'


echo -e "\n\nLogin to the dashboard with token:"
echo -e "\n\thttp://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/#/workloads?namespace=_all#/login\n"
echo -e "\nKubernetes Dashboard Token:\n"
kubectl -n kubernetes-dashboard get secret $(kubectl -n kubernetes-dashboard get sa/admin-user -o jsonpath="{.secrets[0].name}") -o go-template="{{.data.token | base64decode}}"
echo -e "\n\n"

# Create namespace
kubectl create namespace gm-operator

# Create docker pull secrets from local docker config
if [[ "$OSTYPE" == "darwin"* ]]; then
  # Mac OSX
  if [[ -z "$GREYMATTER_REGISTRY_USERNAME" ]]; then
    echo "Docker Username(email): "
    read GREYMATTER_REGISTRY_USERNAME
  fi
  if [[ -z "$GREYMATTER_REGISTRY_PASSWORD" ]]; then
    echo "Docker Password: "
    read -s GREYMATTER_REGISTRY_PASSWORD
  fi
  kubectl create secret docker-registry gm-docker-secret \
    --docker-server=docker.greymatter.io \
    --docker-username=$GREYMATTER_REGISTRY_USERNAME \
    --docker-password=$GREYMATTER_REGISTRY_PASSWORD \
    --docker-email=$GREYMATTER_REGISTRY_USERNAME \
    -n gm-operator
else
  kubectl create secret generic gm-docker-secret \
  --from-file=.dockerconfigjson=$HOME/.docker/config.json \
  --type=kubernetes.io/dockerconfigjson \
  -n gm-operator
fi

# Install GM Operator
kubectl apply -k config/context/kubernetes

while [ "$(kubectl get pods -n gm-operator -l=name='gm-operator' -o jsonpath='{.items[*].status.containerStatuses[0].ready}')" != "true" ]; do
   sleep 5
   echo "Waiting for GM Operator to be ready."
   kubectl get pods -n gm-operator
done
kubectl get pods -n gm-operator

function yes_or_no {
    while true; do
        read -p "$* [y/n]: " yn
        case $yn in
            [Yy]*) return 0  ;;  
            [Nn]*) echo "Aborted" ; return  1 ;;
        esac
    done
}

yes_or_no "Would you like to install a demo mesh?" && echo "
apiVersion: greymatter.io/v1alpha1
kind: Mesh
metadata:
  name: mesh-sample
spec:
  release_version: '1.7'
  zone: default-zone
  install_namespace: default" | kubectl apply -f -


echo -e "\n\n"

export PS1="\[\e[32m\]GM K8S Shell \u@\h \w \n\$ \[\e[m\]"


  '';
}
