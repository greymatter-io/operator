with (import <nixpkgs> {});
mkShell {
  buildInputs = [
    pkgs.coreutils
    pkgs.tmux
    pkgs.kubectl
    pkgs.kubernetes-helm
    pkgs.kompose
    #pkgs.kube3d
    pkgs.kind
    pkgs.kubetail
    pkgs.k9s
  ];

  shellHook = ''
trap 'kind stop cluster; kind delete cluster;' EXIT
docker --version
if docker --version; then
  echo "Docker Daemon is running"
else
  echo "Docker Daemon is not running. Please install and run it on your system."
  exit 0;
fi

echo "🔵 Login to docker.greymatter.io"
docker login docker.greymatter.io

kind create cluster 

kubectl get nodes

#Make sure we are using the right context
kubectl config use-context kind-kind

echo "🔵 Waiting for kind-control-plane to be Ready."
kubectl wait --for=condition=Ready node/kind-control-plane
    
kubectl get nodes


# Create namespace
kubectl create namespace gm-operator
sleep 5;

# Create docker pull secrets from local docker config
if [[ "$OSTYPE" == "darwin"* ]]; then
  # Mac OSX
  if [[ -z "$GREYMATTER_REGISTRY_USERNAME" ]]; then
    echo "🔵 Docker Username(email): "
    read GREYMATTER_REGISTRY_USERNAME
  fi
  if [[ -z "$GREYMATTER_REGISTRY_PASSWORD" ]]; then
    echo "🔵 Docker Password: "
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
function yes_or_no {
    while true; do
        read -p "$* [y/n]: " yn
        case $yn in
            [Yy]*) return 0  ;;  
            [Nn]*) echo "Aborted" ; return  1 ;;
        esac
    done
}

# Install GM Operator
kubectl apply -k config/context/kubernetes


echo "🔵 Waiting for GM Operator to be ready."
while [ "$(kubectl get pods -n gm-operator -l=name='gm-operator' -o jsonpath='{.items[*].status.containerStatuses[0].ready}')" != "true" ]; do
  kubectl get pods -n gm-operator -l=name='gm-operator' -o jsonpath="Name: {.items[0].metadata.name} Status: {.items[0].status.phase}" 2>/dev/null 
  echo -e "\n"
  sleep 5

  #kubectl get pods -n gm-operator
done


kubectl get pods -n gm-operator

yes_or_no "🔵 Would you like to install a demo mesh?" && echo "
apiVersion: greymatter.io/v1alpha1
kind: Mesh
metadata:
  name: mesh-sample
spec:
  release_version: '1.7'
  zone: default-zone
  install_namespace: default" | kubectl apply -f -

echo -e "\n\n"

echo -e "🔵 Available Commands:\n\tk9s\n\tkubectrl\n\tkustomize\n\ttmux\n\thelm\n\tkubetail\n\tkind\n\n"

export PS1="\[\e[32m\]GM K8S Shell \u@\h \w \n\$ \[\e[m\]"


  '';
}
