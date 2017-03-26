/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"fmt"
	"log"

	"github.com/kr/pretty"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	err error
)

func main() {
	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Printf("InClusterConfig Failed: %s\n", err.Error())
		log.Println("Trying local config\n")
	}
	//kubeconfig := flag.String("kubeconfig", "./config", "absolute path to the kubeconfig file")
	//flag.Parse()
	// uses the current context in kubeconfig

	config, err = clientcmd.BuildConfigFromFlags("", "/Users/jk/.kube/config")
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	secret := &v1.Secret{}
	secret.Namespace = "k8s-sslmate"
	secret.Name = "k8s-test"
	secret.Type = v1.SecretTypeTLS
	secret.Data = make(map[string][]byte)
	secret.Data[v1.TLSCertKey] = []byte("thecert123")
	secret.Data[v1.TLSPrivateKeyKey] = []byte("theykey123")

	getSecret, err := clientset.Core().Secrets("k8s-sslmate").Get("k8s-test")
	if err != nil {
		log.Printf("%s\n", err.Error())
		create, err := clientset.Core().Secrets("k8s-sslmate").Create(secret)
		if err != nil {
			log.Print(err.Error())
		}
		//log.Printf("%+v", create)
		fmt.Printf("%# v", pretty.Formatter(create))
	} else {
		//log.Printf("%+v", getSecret)
		fmt.Printf("%# v", pretty.Formatter(getSecret))
		update, err := clientset.Core().Secrets("k8s-sslmate").Update(secret)
		if err != nil {
			log.Print(err.Error())
		}
		//log.Printf("%+v", update)
		fmt.Printf("%# v", pretty.Formatter(update))
	}
}
