package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	//"time"

	"github.com/fsnotify/fsnotify"
	//"github.com/kr/pretty"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	runningConfig *v1.ConfigMap
	configFile    = flag.String("config", "/opt/.kube/config", "Path to kubectl config file")
	certDir       = flag.String("certdir", "/etc/sslmate/", "Specify the SSLmate cert directory")
	keyDir        = flag.String("keydir", "/etc/sslmate/keys/", "Specify the SSLmate key directory")
)

const (
	VERSION = "0.2-springroll"
)

func init() {
	flag.Parse()
	*certDir = path.Clean(*certDir)
	*keyDir = path.Clean(*keyDir)

	if cdir, err := isDir(*certDir); err != nil || cdir == false {
		log.Fatalf("Cert dir \"%s\" is not a valid directory", *certDir)
	}

	if kdir, err := isDir(*keyDir); err != nil || kdir == false {
		log.Fatalf("Key dir \"%s\" is not a valid directory", *keyDir)
	}
}

func isDir(pth string) (bool, error) {
	fi, err := os.Stat(pth)
	if err != nil {
		return false, err
	}
	return fi.IsDir(), nil
}

func deploySecret(clientset *kubernetes.Clientset, namespace_in string, secretObj *v1.Secret) {
	namespace := strings.TrimSpace(namespace_in)
	if updateSecret(clientset, namespace, secretObj) {
		log.Printf("INFO: Secret \"%s/%s\" updated", namespace, secretObj.ObjectMeta.Name)
	} else {
		if createSecret(clientset, namespace, secretObj) {
			log.Printf("INFO: Secret \"%s/%s\" created", namespace, secretObj.ObjectMeta.Name)
		} else {
			log.Printf("ERROR: Update & Create failed for \"%s/%s\"", namespace, secretObj.ObjectMeta.Name)
		}
	}
}

func createSecret(clientset *kubernetes.Clientset, namespace string, secretObj *v1.Secret) bool {
	_, err := clientset.Core().Secrets(namespace).Create(secretObj)
	if err == nil {
		return true
	} else {
		log.Printf("ERROR Creating: \"%s/%s\": %s", namespace, secretObj.Name, err)
		return false
	}
}

func updateSecret(clientset *kubernetes.Clientset, namespace string, secretObj *v1.Secret) bool {
	_, err := clientset.Core().Secrets(namespace).Update(secretObj)
	if err == nil {
		return true
	} else {
		log.Printf("ERROR Updating \"%s/%s\": %s", namespace, secretObj.Name, err)
		return false
	}
}

func updateCert(clientset *kubernetes.Clientset, domainname string) bool {

	tlscrt_f, cert_err := ioutil.ReadFile(fmt.Sprintf("%s/%s.chained.crt", *certDir, domainname))
	if cert_err != nil {
		log.Printf("CERT ERROR: %s", cert_err)
		return false
	}

	tlskey_f, key_err := ioutil.ReadFile(fmt.Sprintf("%s/%s.key", *keyDir, domainname))
	if key_err != nil {
		log.Printf("KEY ERROR: %s", key_err)
		return false
	}

	// fmt.Printf("%# v", pretty.Formatter(secret))
	if cert_err == nil && key_err == nil {
		if val, ok := runningConfig.Data[domainname]; ok {
			log.Printf("INFO: \"%s\" found in configMap with namespace(s) [\"%s\"]\n", domainname, val)
			namespaces := strings.Split(val, ",")
			for _, namespace := range namespaces {
				secret := &v1.Secret{}
				secret.Namespace = namespace
				secret.Name = domainname
				secret.Type = v1.SecretTypeTLS
				secret.Data = make(map[string][]byte)
				secret.Data[v1.TLSCertKey] = tlscrt_f
				secret.Data[v1.TLSPrivateKeyKey] = tlskey_f
				deploySecret(clientset, namespace, secret)
				//fmt.Printf("%# v\n", pretty.Formatter(secret))
			}
		}
	}

	return true
}

func main() {
	log.Printf("Starting k8s-sslmate %s\n", VERSION)
	log.Printf("Cert Path:\t[\"%s\"]", *certDir)
	log.Printf("Key Path:\t[\"%s\"]", *keyDir)

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("WARN: InClusterConfig Failed: %s\n", err.Error())
	}

	//Build a clientcmd client if in-cluster config fails
	if config == nil {
		log.Println("INFO: Trying local config")
		config, err = clientcmd.BuildConfigFromFlags("", *configFile)
		if err != nil {
			log.Fatalf("FATAL: initializing clientcmd failed \"%s\"\n", err.Error())
		} else {
			log.Printf("INFO: Local config succeeded, using clientcmd [\"%s\"]\n", *configFile)
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("FATAL: Could not create new client %s", err.Error())
	}

	getConfigmap(clientset, true)

	startCron(clientset)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				// log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					if strings.HasSuffix(event.Name, ".chained.crt") {
						//log.Println("modified file:", event.Name)
						updateCert(clientset, path.Base(strings.TrimSuffix(event.Name, ".chained.crt")))
					}
				}
			case watch_err := <-watcher.Errors:
				log.Println("ERROR: ", watch_err)
			}
		}
	}()

	err = watcher.Add(*certDir)
	if err != nil {
		log.Fatal(err)
	}
	run_SSLmate()
	<-done

}
