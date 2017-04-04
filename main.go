package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	//"github.com/kr/pretty"
	"github.com/robfig/cron"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	runningConfig *v1.ConfigMap
)

func getConfigmap(clientset *kubernetes.Clientset, firstrun bool) {
	oldConfigMap := runningConfig
	cm, err := clientset.Core().ConfigMaps("k8s-sslmate").Get("k8s-sslmate-config")
	if err != nil {
		log.Fatalf("FATAL: Can't get configmap!, %s", err.Error())
	}
	runningConfig = cm

	if firstrun == false {
		// Check if the last known configmap is the same
		if reflect.DeepEqual(oldConfigMap, cm) == false {
			log.Print("INFO: Configmap updated, propagating changes")
			for k := range cm.Data {
				if reflect.DeepEqual(oldConfigMap.Data[k], cm.Data[k]) == false {
					updateCert(clientset, k)
				}
			}
		} //End check if configmap the same
	}
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

	tlscrt_f, err := ioutil.ReadFile(fmt.Sprintf("/etc/sslmate/%s.chained.crt", domainname))
	if err != nil {
		log.Printf("ERROR: %s", err)
		return false
	}

	tlskey_f, err := ioutil.ReadFile(fmt.Sprintf("/etc/sslmate/keys/%s.key", domainname))
	if err != nil {
		log.Printf("ERROR: %s", err)
		return false
	}

	// fmt.Printf("%# v", pretty.Formatter(secret))

	if val, ok := runningConfig.Data[domainname]; ok {
		log.Printf("INFO: \"%s\" found in ConfigMap with namespace(s) [\"%s\"]\n", domainname, val)
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

	return true
}

func run_SSLmate() bool {
	cmd := exec.Command("sslmate", "download", "--all")
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("INFO: Waiting for SSLmate to finish...")
	err = cmd.Wait()
	if err != nil {
		//fmt.Printf("%# v\n", pretty.Formatter(err))
		log.Print("INFO: SSLmate had no new certs to download")
	}
	if err == nil {
		log.Print("INFO: SSLmate has downloaded new certs")
	}
	return true
}

func delayedStart() {
	time.Sleep(2 * time.Second)
	run_SSLmate()
}

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("WARN: InClusterConfig Failed: %s\n", err.Error())
	}

	if config == nil {
		log.Println("INFO: Trying local config\n")
		config, err = clientcmd.BuildConfigFromFlags("", "/opt/.kube/config")
		if err != nil {
			panic(err.Error())
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("FATAL: Could not create new client %s", err.Error())
	}

	getConfigmap(clientset, true)

	c := cron.New()
	c.AddFunc("@every 60m", func() { run_SSLmate() })
	c.AddFunc("@every 1m", func() { getConfigmap(clientset, false) })
	c.Start()
	go delayedStart()
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
			case err := <-watcher.Errors:
				log.Println("ERROR: ", err)
			}
		}
	}()

	err = watcher.Add("/etc/sslmate")
	if err != nil {
		log.Fatal(err)
	}
	//time.Sleep(100 * time.Second)
	<-done

}
