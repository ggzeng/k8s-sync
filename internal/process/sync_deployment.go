package process

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
	"k8sync/internal/config"
	k8client "k8sync/internal/k8s/client"
	"k8sync/pkg/logger"
	"os"
)

func SyncDeployment(srcK8 *k8client.K8s, dstK8 *k8client.K8s) error {
	var err error
	var srcList *appsv1.DeploymentList
	var dstList *appsv1.DeploymentList

	srcDeployClient := srcK8.Clientset.AppsV1().Deployments(srcK8.GetNamespace())
	srcList, err = srcDeployClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	dstDeployClient := dstK8.Clientset.AppsV1().Deployments(dstK8.GetNamespace())
	dstList, err = dstDeployClient.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	/* save destination deployment */
	var ddMap = make(map[string]appsv1.Deployment)
	var dcMap = make(map[string]corev1.Container)
	for _, dd := range dstList.Items {
		deployFilter(&dd)
		ddMap[dd.Name] = dd
		for _, dc := range dd.Spec.Template.Spec.Containers {
			dcMap[dd.Name+"-"+dc.Name] = dc
		}
	}

	/* compare source and destination deployment */
	logger.Infof("sync deployment")
	for _, sd := range srcList.Items {
		deployFilter(&sd)
		if config.GetBool("app.yaml") {
			exportDeployYaml(srcK8.GetNamespace(), &sd)
		}
		if _, ok := ddMap[sd.Name]; !ok {
			logger.Infof("  create deployment: %s", sd.Name)
			_, err = dstDeployClient.Create(context.TODO(), &sd, metav1.CreateOptions{})
			if err != nil {
				return err
			}
		} else {
			logger.Infof("  update deployment: %s", sd.Name)
			for _, sc := range sd.Spec.Template.Spec.Containers {
				if _, ok := dcMap[sd.Name+"-"+sc.Name]; !ok {
					logger.Infof("    add container: %s", sc.Name)
				}else {
					if sc.Image != dcMap[sd.Name+"-"+sc.Name].Image {
						logger.Infof("    update container: %s's image: %s", sc.Name, sc.Image)
					}
				}
			}
			_, err = dstDeployClient.Update(context.TODO(), &sd, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
			delete(ddMap, sd.Name)
		}
	}

	/* delete destination deployment */
	for _, dd := range ddMap {
		logger.Infof("  delete deployment: %s", dd.Name)
		err = dstDeployClient.Delete(context.TODO(), dd.Name, metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}

	return nil
}

func deployFilter(d *appsv1.Deployment) {
	d.Namespace = ""
	d.CreationTimestamp = metav1.Time{}
	d.Status = appsv1.DeploymentStatus{}
	d.ManagedFields = []metav1.ManagedFieldsEntry{}
	d.UID = ""
	d.ResourceVersion = ""
}

func exportDeployYaml(ns string, d *appsv1.Deployment)  {

	path := "yaml/"+ns
	err := os.MkdirAll(path, 0750)
	if err != nil && !os.IsExist(err) {
		logger.Fatal(err)
	}
	newFile, err := os.Create(path+"/"+d.Name+"-deploy.yaml")
	if err != nil {
		logger.Fatal(err)
	}
	defer newFile.Close()

	y := printers.YAMLPrinter{}
	d.Kind = "Deployment"
	d.APIVersion = "apps/v1"
	delete(d.Annotations, "kubectl.kubernetes.io/last-applied-configuration")
	delete(d.Annotations, "deployment.kubernetes.io/revision")
	if err := y.PrintObj(d, newFile); err != nil {
		logger.Fatal(err)
	}
}
