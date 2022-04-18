package cmd

import (
	"github.com/spf13/cobra"
	"k8sync/internal/config"
	k8client "k8sync/internal/k8s/client"
	"k8sync/internal/process"
	"k8sync/pkg/logger"
)

func cliStart(cmd *cobra.Command, args []string) {
	srcNamesapce := config.GetString("src.namespace")
	if srcNamesapce == "" {
		logger.Fatal("src.namespace is empty")
	}
	dstNamesapce := config.GetString("dst.namespace")
	if dstNamesapce == "" {
		dstNamesapce = srcNamesapce
	}
	srcK8 := k8client.New("src")
	srcK8.SetNamespace(srcNamesapce)
	logger.Infof("from src namespace: %s", srcNamesapce)
	dstK8 := k8client.New("dst")
	dstK8.SetNamespace(dstNamesapce)
	logger.Infof("to  dest namespace: %s", dstNamesapce)
	objs := config.GetStringSlice("src.objects")

	for _, obj := range objs {
		switch obj {
		case "service":
			if err := process.SyncService(srcK8, dstK8); err != nil {
				logger.Fatal(err)
			}
		case "deployment":
			if err := process.SyncDeployment(srcK8, dstK8); err != nil {
				logger.Fatal(err)
			}
		}
	}
}
