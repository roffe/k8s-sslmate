# SSLmate to automatic sync certs with K8S cluster


dep init
dep ensure k8s.io/client-go@^2.0.0

curl -v --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes/



curl https://$KUBERNETES_PORT_443_TCP_ADDR/api/v1/namespaces/k8s-sslmate/secrets \
-s --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
-H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
-XPOST -H"Content-Type: application/yaml" -d@<(update_certs.sh)