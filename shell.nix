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
docker --version
if docker --version; then
  echo "Docker Daemon is running" 
else
  echo "Docker Daemon is not running. Please install and run it on your system." 
  exit 0;
fi

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


# Create namespace
kubectl create namespace gm-operator

# Create docker pull secrets from local docker config
docker login docker.greymatter.io
    kubectl create secret generic gm-docker-secret \
    --from-file=.dockerconfigjson=$HOME/.docker/config.json \
    --type=kubernetes.io/dockerconfigjson \
    -n gm-operator

# Install GM Operator
kubectl apply -k config/context/kubernetes

kubectl create secret generic regcred \
   --from-file=.dockerconfigjson=$HOME/.docker/config.json \
   --type=kubernetes.io/dockerconfigjson \
   -n default

echo "Just a moment..."
sleep 5;
echo "Adding private credentials to service account."
kubectl patch serviceaccount default -p '{"imagePullSecrets": [{"name": "regcred"}]}'
 
echo -e "\n\nLogin to the dashboard with token:"
echo -e "\n\thttp://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/\n"
echo -e "\nKubernetes Dashboard Token:\n"
kubectl -n kubernetes-dashboard get secret $(kubectl -n kubernetes-dashboard get sa/admin-user -o jsonpath="{.secrets[0].name}") -o go-template="{{.data.token | base64decode}}"

echo -e "\n\n"

export PS1="\[\e[32m\]GM K8S Shell \u@\h \w \n\$ \[\e[m\]"


trap 'k3d cluster stop gm-cluster; k3d cluster delete gm-cluster;  killall kubectl;' EXIT
  '';
}
