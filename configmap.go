package main

import (
	"k8s.io/client-go/kubernetes"
	"log"
	"reflect"
)

func getConfigmap(clientset *kubernetes.Clientset, firstrun bool) {
	oldConfigMap := runningConfig
	cm, err := clientset.Core().ConfigMaps("k8s-sslmate").Get("k8s-sslmate-config")
	if err != nil {
		log.Fatalf("FATAL: Can't get configMap!, %s", err.Error())
	}
	runningConfig = cm

	if firstrun == false {
		// Check if the last known configmap is the same
		if reflect.DeepEqual(oldConfigMap, cm) == false {
			log.Print("INFO: configMap updated, propagating changes")
			for k := range cm.Data {
				if reflect.DeepEqual(oldConfigMap.Data[k], cm.Data[k]) == false {
					updateCert(clientset, k)
				}
			}
		} //End check if configmap the same
	}
}
