package main

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	config, err := clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	podInformer := factory.Core().V1().Pods()

	podInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldPod := oldObj.(*corev1.Pod)
			newPod := newObj.(*corev1.Pod)

			for i, newCS := range newPod.Status.ContainerStatuses {
				if newCS.State.Terminated == nil || newCS.State.Terminated.Reason != "OOMKilled" {
					continue
				}

				if i < len(oldPod.Status.ContainerStatuses) {
					oldCS := oldPod.Status.ContainerStatuses[i]
					if oldCS.State.Terminated != nil && oldCS.State.Terminated.Reason == "OOMKilled" {
						continue
					}
				}

				fmt.Printf("[报警] UPDATE Namespace %s 的 Pod %s 发生 OOMKilled！\n", newPod.Namespace, newPod.Name)
			}
		},
		AddFunc: func(obj interface{}) {
			pod := obj.(*corev1.Pod)
			for _, cs := range pod.Status.ContainerStatuses {
				if cs.State.Terminated != nil && cs.State.Terminated.Reason == "OOMKilled" {
					fmt.Printf("[报警] ADD Namespace %s 的 Pod %s 发生 OOMKilled！\n", pod.Namespace, pod.Name)
				} else if cs.LastTerminationState.Terminated != nil && cs.LastTerminationState.Terminated.Reason == "OOMKilled" {
					fmt.Printf("[报警] ADD Namespace %s 的 Pod %s 发生 OOMKilled！\n", pod.Namespace, pod.Name)
				}
			}
		},
	})

	factory.Start(make(chan struct{}))
	select {}
}
