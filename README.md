# SSLmate to automatic sync certs with K8S cluster


dep init
dep ensure k8s.io/client-go@^2.0.0

curl -v --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://kubernetes/
