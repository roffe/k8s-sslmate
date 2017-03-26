cat > /update_certs.sh <<"EORT"
#!bin/sh

for CERTNAME in `find /etc/sslmate/ -maxdepth 1 -type f -name '*chained*'`; do
 CERTHOST=$(echo ${CERTNAME}|rev|cut -d'.' -f3- |rev|cut -d'/' -f4-)
 cat /secret.tmpl | \
 sed -e "s/%TLS_CRT%/$(cat $CERTNAME | base64 -w 0)/g" \
     -e "s/%TLS_KEY%/$(cat /etc/sslmate/keys/$CERTHOST.key | base64 -w 0)/g" \
     -e "s/%DOMAIN%/$CERTHOST/g" \
     -e "s/%NAMESPACE%/k8s-sslmate/g" > out.json
curl https://$KUBERNETES_PORT_443_TCP_ADDR/api/v1/namespaces/k8s-sslmate/secrets \
-s --cacert /var/run/secrets/kubernetes.io/serviceaccount/ca.crt \
-H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" \
-XPOST -H"Content-Type: application/yaml" -d@out.json
done
EORT