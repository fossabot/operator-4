package mainhandler

import (
	"fmt"
	"k8s-ca-websocket/cautils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/armosec/capacketsgo/apis"
	pkgcautils "github.com/armosec/capacketsgo/cautils"
	"github.com/armosec/capacketsgo/k8sinterface"
	"github.com/armosec/capacketsgo/secrethandling"
)

var IgnoreCommandInNamespace = map[string][]string{}

func InitIgnoreCommandInNamespace() {
	if len(IgnoreCommandInNamespace) != 0 {
		return
	}
	IgnoreCommandInNamespace[apis.UPDATE] = []string{metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE}
	IgnoreCommandInNamespace[apis.INJECT] = []string{metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE}
	IgnoreCommandInNamespace[apis.DECRYPT] = []string{metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE}
	IgnoreCommandInNamespace[apis.ENCRYPT] = []string{metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE}
	IgnoreCommandInNamespace[apis.REMOVE] = []string{metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE}
	IgnoreCommandInNamespace[apis.RESTART] = []string{metav1.NamespaceSystem, metav1.NamespacePublic}
	IgnoreCommandInNamespace[apis.SCAN] = []string{}
	// apis.SCAN:    {metav1.NamespaceSystem, metav1.NamespacePublic, cautils.CA_NAMESPACE},
}

func ignoreNamespace(command, namespace string) bool {
	InitIgnoreCommandInNamespace()
	if s, ok := IgnoreCommandInNamespace[command]; ok {
		for i := range s {
			if s[i] == namespace {
				return true
			}
		}
	}
	return false
}
func (mainHandler *MainHandler) listWorkloads(namespace, resource string, labels map[string]string) ([]k8sinterface.Workload, error) {
	groupVersionResource, err := k8sinterface.GetGroupVersionResource(resource)
	if err != nil {
		return nil, err
	}
	return mainHandler.k8sAPI.ListWorkloads(&groupVersionResource, namespace, labels)
}
func (mainHandler *MainHandler) GetResourcesIDs(namespace string, workloads []k8sinterface.Workload) ([]string, []error) {
	errs := []error{}
	idMap := make(map[string]interface{})
	for i := range workloads {
		switch workloads[i].GetKind() {
		case "Namespace":
			idMap[pkgcautils.GetWLID(cautils.CA_CLUSTER_NAME, workloads[i].GetName(), "namespace", workloads[i].GetName())] = true
		case "Secret":
			// check if secret type supported
			// check is shadow secret
			idMap[secrethandling.GetSID(cautils.CA_CLUSTER_NAME, namespace, workloads[i].GetName(), "")] = true
		default:
			if wlid := workloads[i].GetWlid(); wlid != "" {
				idMap[wlid] = true
			} else {
				// find wlid
				kind, name, err := mainHandler.k8sAPI.CalculateWorkloadParentRecursive(&workloads[i])
				if err != nil {
					errs = append(errs, fmt.Errorf("CalculateWorkloadParentRecursive: namespace: %s, pod name: %s, error: %s", workloads[i].GetNamespace(), workloads[i].GetName(), err.Error()))
				}
				wlid := pkgcautils.GetWLID(cautils.CA_CLUSTER_NAME, namespace, kind, name)
				if wlid != "" {
					idMap[wlid] = true
				}
			}
		}
	}
	return cautils.MapToString(idMap), errs
}

func getCommandNamespace(command *apis.Command) string {
	if command.Wlid != "" {
		return pkgcautils.GetNamespaceFromWlid(command.Wlid)
	}
	if command.WildWlid != "" {
		return cautils.GetNamespaceFromWildWlid(command.WildWlid)
	}
	return ""
}

func getCommandID(command *apis.Command) string {
	if command.Wlid != "" {
		return command.Wlid
	}
	if command.WildWlid != "" {
		return command.WildWlid
	}
	return ""
}

func resourceList(command string) []string {
	switch command {
	case apis.UNREGISTERED:
		return []string{"namespaces", "pods"}
		// return []string{"namespaces", "secrets", "pods"}
	case apis.DECRYPT, apis.ENCRYPT:
		return []string{"secrets"}
	default:
		return []string{"pods"}

	}

}
