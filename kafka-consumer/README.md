export KO_DOCKER_REPO=ksgregistry.azurecr.io/kafka-consumer
az acr login -n ksgregistry.azurecr.io
ko build .
kubectl exec -it $pod -c $container -- bash