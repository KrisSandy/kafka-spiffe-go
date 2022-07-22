export KO_DOCKER_REPO=ksgregistry.azurecr.io/kafka-producer
az acr login -n ksgregistry.azurecr.io
ko build .