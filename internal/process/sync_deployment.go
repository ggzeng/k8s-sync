package process

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8client "k8sync/internal/k8s/client"
)

func SyncDeployment(srcK8 *k8client.K8s, dstK8 *k8client.K8s, srcNs string, dstNs string) error {
	client := srcK8.Clientset.AppsV1().Deployments(srcK8.GetNamespace())
	list, err := client.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, d := range list.Items {
		fmt.Printf(" * %s (%d replicas)\n", d.Name, *d.Spec.Replicas)
	}

	// -- 基于源 --
	// 创建对应k8s client
	// 获取此client所有的object

	// -- 基于目的 --
	// 创建对应k8s client
	// 获取此client所有的object, 并给予name保存到map中

	// 遍历源中的object
	// 看是否在目的中有，没有则创建，有则更新
	// 如果是更新的话，更新完之后，将其从目的map中删除
	return nil
}
