package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func test(cs *kubernetes.Clientset) {
	_, err := cs.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Println("Error happened")
		panic(err.Error())
	}

	fmt.Println("Testing... Authentication to API server succeeded")
}
