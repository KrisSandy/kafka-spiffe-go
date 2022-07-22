kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://nostak.org/ns/kafka/sa/default \
    -parentID spiffe://nostak.org/ns/spire/sa/spire-agent \
    -selector k8s:ns:kafka \
    -selector k8s:sa:default

kubectl exec -n spire spire-server-0 -- \
    /opt/spire/bin/spire-server entry create \
    -spiffeID spiffe://nostak.org/ns/unauth/sa/default \
    -parentID spiffe://nostak.org/ns/spire/sa/spire-agent \
    -selector k8s:ns:unauth \
    -selector k8s:sa:default    