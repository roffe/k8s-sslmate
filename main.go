package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"path"
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

func getConfigmap(clientset *kubernetes.Clientset) {
	cm, err := clientset.Core().ConfigMaps("k8s-sslmate").Get("k8s-sslmate-config")
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("%# v\n", pretty.Formatter(cm.Data))
	runningConfig = cm
}

func deploySecret(clientset *kubernetes.Clientset, namespace_in string, secretObj *v1.Secret) {
	namespace := strings.TrimSpace(namespace_in)
	if updateSecret(clientset, namespace, secretObj) {
		log.Printf("Secret %s/%s updated", namespace, secretObj.ObjectMeta.Name)
	} else {
		if createSecret(clientset, namespace, secretObj) {
			log.Printf("Secret %s/%s created", namespace, secretObj.ObjectMeta.Name)
		} else {
			log.Printf("Update & Create failed for %s/%s", namespace, secretObj.ObjectMeta.Name)
		}
	}

}

func createSecret(clientset *kubernetes.Clientset, namespace string, secretObj *v1.Secret) bool {
	_, err := clientset.Core().Secrets(namespace).Create(secretObj)
	if err == nil {
		return true
	} else {
		log.Printf("WARN Creating: %s", err)
		return false
	}
}

func updateSecret(clientset *kubernetes.Clientset, namespace string, secretObj *v1.Secret) bool {
	_, err := clientset.Core().Secrets(namespace).Update(secretObj)
	if err == nil {
		return true
	} else {
		log.Printf("ERROR Updating: %s", err)
		return false
	}
}

func updateCert(clientset *kubernetes.Clientset, tlscrt string) bool {
	domainname := path.Base(strings.TrimSuffix(tlscrt, ".chained.crt"))
	tlscrt_f, err := ioutil.ReadFile(tlscrt)
	if err != nil {
		log.Panic(err)
	}
	tlskey_f, err := ioutil.ReadFile(fmt.Sprintf("/etc/sslmate/keys/%s.key", domainname))
	if err != nil {
		log.Panic(err)
	}

	// fmt.Printf("%# v", pretty.Formatter(secret))

	if val, ok := runningConfig.Data[domainname]; ok {
		log.Printf("%s found in ConfigMap with namespaces [\"%s\"]\n", domainname, val)
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
	log.Printf("Waiting for SSLmate to finish...")
	err = cmd.Wait()
	if err != nil {
		//fmt.Printf("%# v\n", pretty.Formatter(err))
		log.Print("SSLmate had no new certs to download")
	}
	if err == nil {
		log.Print("SSLmate has downloaded new certs")
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
		log.Printf("InClusterConfig Failed: %s\n", err.Error())
	}

	if config == nil {
		log.Println("Trying local config\n")
		config, err = clientcmd.BuildConfigFromFlags("", "/opt/.kube/config")
		if err != nil {
			panic(err.Error())
		}
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("Could not create new client %s", err.Error())
	}

	getConfigmap(clientset)

	c := cron.New()
	c.AddFunc("@every 60m", func() { run_SSLmate() })
	c.AddFunc("@every 1m", func() { getConfigmap(clientset) })
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
						log.Println("modified file:", event.Name)
						updateCert(clientset, event.Name)
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
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
