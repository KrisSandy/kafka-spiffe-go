docker build -t wk-kafka .
docker tag wk-kafka ksgregistry.azurecr.io/wk-kafka
az login
az acr login -n ksgregistry.azurecr.io
docker push ksgregistry.azurecr.io/wk-kafka
docker run -it wk-kafka sh