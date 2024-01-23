package main

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func test() {
	_, err := clientSet.CoreV1().Pods("kube-system").List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Println("Error happened")
		panic(err.Error())
	}

	fmt.Println("Authentication to API server succeeded")
}
