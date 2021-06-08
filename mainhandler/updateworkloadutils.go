package mainhandler

import (
	"context"
	"k8s-ca-websocket/cautils"

	"github.com/armosec/capacketsgo/k8sinterface"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func isForceDelete(args map[string]interface{}) bool {
	if args == nil || len(args) == 0 {
		return false
	}
	if v, ok := args["forceDelete"]; ok && v != nil {
		return v.(bool)
	}
	return false
}

func (actionHandler *ActionHandler) deleteConfigMaps() error {
	confName := cautils.GenarateConfigMapName(actionHandler.wlid)
	return actionHandler.k8sAPI.KubernetesClient.CoreV1().ConfigMaps(cautils.GetNamespaceFromWlid(actionHandler.wlid)).Delete(context.Background(), confName, metav1.DeleteOptions{})
}

func (actionHandler *ActionHandler) deleteWorkloadTemplate() error {
	return actionHandler.cacli.WTDelete(actionHandler.wlid)
}

func persistentVolumeFound(workload *k8sinterface.Workload) bool {
	volumes, _ := workload.GetVolumes()
	for _, vol := range volumes {
		if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName != "" {
			return true
		}
	}
	return false
}
