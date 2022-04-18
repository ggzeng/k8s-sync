package process

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8sync/internal/config"
	k8client "k8sync/internal/k8s/client"
	"k8sync/pkg/logger"
	"os"
)

func SyncService(srcK8 *k8client.K8s, dstK8 *k8client.K8s) error {
	var err error
	var srcList *corev1.ServiceList
	var dstList *corev1.ServiceList

	srcServiceClient := srcK8.Clientset.CoreV1().Services(srcK8.GetNamespace())
	srcList, err = srcServiceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	dstServiceClient := dstK8.Clientset.CoreV1().Services(dstK8.GetNamespace())
	dstList, err = dstServiceClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	/* save destination deployment */
	var dsMap = make(map[string]corev1.Service)     // destination service map
	var dpMap = make(map[string]corev1.ServicePort) // destination service port map
	for _, ds := range dstList.Items {
		dsMap[ds.Name] = ds
		for _, dp := range ds.Spec.Ports {
			dpMap[ds.Name+"-"+dp.Name] = dp
		}
	}

	/* compare source and destination deployment */
	logger.Infof("sync service")
	for _, ss := range srcList.Items {
		serviceFilter(&ss)
		if config.GetBool("app.yaml") {
			exportServiceYaml(srcK8.GetNamespace(), &ss)
		}
		if _, ok := dsMap[ss.Name]; !ok {
			logger.Infof("  create service: %s", ss.Name)
			_, err = dstServiceClient.Create(context.TODO(), &ss, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			logger.Infof("  update service: %s", ss.Name)
			for _, sp := range ss.Spec.Ports {
				if _, ok := dpMap[ss.Name+"-"+sp.Name]; !ok {
					logger.Infof("    add port: %s", ss.Name)
				} else {
					if sp.Port != dpMap[ss.Name+"-"+sp.Name].Port {
						logger.Infof("    change port: %s from %d to %d",
							sp.Name, dpMap[ss.Name+"-"+sp.Name].Port, sp.Port)
					}
				}
			}
			ds := dsMap[ss.Name]
			ds.Spec.Ports = ss.Spec.Ports
			_, err = dstServiceClient.Update(context.TODO(), &ds, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			delete(dsMap, ss.Name)
		}
	}

	/* delete destination deployment */
	for _, ds := range dsMap {
		logger.Infof("  delete service: %s", ds.Name)
		err = dstServiceClient.Delete(context.TODO(), ds.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func serviceFilter(s *corev1.Service) {
	s.Namespace = ""
	s.CreationTimestamp = metav1.Time{}
	s.Status = corev1.ServiceStatus{}
	s.ManagedFields = []metav1.ManagedFieldsEntry{}
	s.UID = ""
	s.ResourceVersion = ""
	s.Spec.ClusterIP = ""
	s.Spec.ClusterIPs = nil
}

func exportServiceYaml(ns string, s *corev1.Service) {

	path := "yaml/" + ns
	err := os.MkdirAll(path, 0750)
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}
	newFile, err := os.Create(path + "/" + s.Name + "-service.yaml")
	if err != nil {
		logger.Fatal(err)
	}
	defer newFile.Close()

	y := printers.YAMLPrinter{}
	s.Kind = "Service"
	s.APIVersion = "v1"
	delete(s.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	delete(s.Annotations, "deployment.kubernetes.io/revision")
	if err := y.PrintObj(s, newFile); err != nil {
		logger.Fatal(err)
	}
}
