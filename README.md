# k8s-sslmate
Your buddy to keep sslmate certificates synchronized with your Kubernetes cluster

*Reloads it's internal config map every 1 minute  
*Checks for new SSLmate certificates to download at start & every 60 minutes  

At startup all certs are downloaded and pushed / created according to mappings in configmap

If configmap is updated changes will be propagated within 1 minute

If new SSL certs are added to the privatekey configMap they will be added at the next run ( every 60 minute )
This can be speedup by removing the pod and thereby forcing a complete propagation of all certs.


## Building
dep init 
dep ensure k8s.io/client-go@^2.0.0  

## Local testing
When started in a local docker the K8S clientcmd package is used and will need a config file containing certs / token to talk to a K8S cluster
```
docker run --rm -it --name k8s-sslmate -e SSLMATE_API_KEY="YourSSLmateAPIkey" -v /path/to/.kube:/opt/.kube roffe/k8s-sslmate
```

## Deployment to K8S
There are deployment manifests included in this repo:

### Preparations
**Attention!: k8s-sslmate assumes that the lowercase word 'star' is used for wildcard certificates and will configure SSLmate to act accordingly!**

To create a secret containing your privatekeys used with SSLmate issue the following after creating the namespace

```
kubectl create secret generic sslmate-private-keys --from-file=domain.tld.key --from-file=star.somedomain.tld.key --namespace k8s-sslmate
```

#### 00-namespace.yaml
Creates the namespace **k8s-sslmate** where the application will be running
```
kubectl create -f manifests/00-namespace.yaml
````

#### 01-configmap.yaml
Edit to suit your needs. The mapping is very simple where the domain name is the key and a comma separated list after is the namespaces to deploy the CERTs to.
```
kubectl create -f manifests/01-configmap.yaml
```

#### 02-sslmate-api-key
Base64 encode your SSLmate API key and insert into the template. then create with  
```
kubectl create -f manifests/02-sslmate-api-key.yaml
```

#### 03-deployment.yaml
The actuall deployment. It will reference your sslmate-api-key secret and use as a environment variable
