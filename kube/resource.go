package kube

import (
	"fmt"
	"k8s.io/api/autoscaling/v2beta2"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	prompt "github.com/c-bata/go-prompt"
	"github.com/c-bata/kube-prompt/internal/debug"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Suggests []prompt.Suggest

/* Suggest sort interface 主要把cronjob排在后面，否则太影响效率了 */

func (s Suggests) Len() int {
	return len(s)
}

func (s Suggests) Less(i, j int) bool {
	iB := 0
	jB := 0
	if strings.Contains(s[i].Text, "cronjob") {
		iB = 1
	}
	if strings.Contains(s[j].Text, "cronjob") {
		jB = 1
	}
	if iB != jB {
		return iB <= jB
	}

	return s[i].Text <= s[j].Text
}

func (s Suggests) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

/* resource init */

const thresholdFetchInterval = 10 * time.Second

func init() {
	lastFetchedAt = new(sync.Map)
	podList = new(sync.Map)
	endpointList = new(sync.Map)
	deploymentList = new(sync.Map)
	daemonSetList = new(sync.Map)
	eventList = new(sync.Map)
	secretList = new(sync.Map)
	ingressList = new(sync.Map)
	limitRangeList = new(sync.Map)
	persistentVolumeClaimsList = new(sync.Map)
	podTemplateList = new(sync.Map)
	replicaSetList = new(sync.Map)
	replicationControllerList = new(sync.Map)
	resourceQuotaList = new(sync.Map)
	serviceAccountList = new(sync.Map)
	serviceList = new(sync.Map)
	jobList = new(sync.Map)
	cronJobList = new(sync.Map)
	hpaList = new(sync.Map)
}

/* LastFetchedAt */

var (
	lastFetchedAt *sync.Map
)

func FlushAllResourceListCache() {
	lastFetchedAt = new(sync.Map)
}

/**
十秒钟内不会重新拉新的pod信息
*/
func shouldFetch(key string) bool {
	v, ok := lastFetchedAt.Load(key)
	if !ok {
		return true
	}
	t, ok := v.(time.Time)
	if !ok {
		return true
	}
	return time.Since(t) > thresholdFetchInterval
}

func updateLastFetchedAt(key string) {
	lastFetchedAt.Store(key, time.Now())
}

/* Component Status */

var (
	componentStatusList atomic.Value
)

func fetchComponentStatusList(client *kubernetes.Clientset) {
	key := "component_status"
	if !shouldFetch(key) {
		return
	}
	l, _ := client.CoreV1().ComponentStatuses().List(metav1.ListOptions{})
	componentStatusList.Store(l)
	updateLastFetchedAt(key)
}

func getComponentStatusCompletions(client *kubernetes.Clientset) []prompt.Suggest {
	go fetchComponentStatusList(client)
	l, ok := componentStatusList.Load().(*corev1.ComponentStatusList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Config Maps */

var (
	configMapsList atomic.Value
)

func fetchConfigMapList(client *kubernetes.Clientset, namespace string) {
	key := "config_map_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)
	l, _ := client.CoreV1().ConfigMaps(namespace).List(metav1.ListOptions{})
	configMapsList.Store(l)
}

func getConfigMapSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchConfigMapList(client, namespace)
	l, ok := configMapsList.Load().(*corev1.ConfigMapList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Contexts */

var (
	contextList atomic.Value
)

func fetchContextList() {
	key := "context"
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)
	r := ExecuteAndGetResult("config get-contexts --no-headers -o name")
	r = strings.TrimRight(r, "\n")
	contextList.Store(strings.Split(r, "\n"))
}

func getContextSuggestions() []prompt.Suggest {
	go fetchContextList()
	l, ok := contextList.Load().([]string)
	if !ok || len(l) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l))
	for i := range l {
		s[i] = prompt.Suggest{
			Text: l[i],
		}
	}
	sort.Sort(s)
	return s
}

/* Pod */

var (
	podList *sync.Map
)

func fetchPods(client *kubernetes.Clientset, namespace string) {
	key := "pod_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Pods(namespace).List(metav1.ListOptions{})
	podList.Store(namespace, l)
}

func getPodSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchPods(client, namespace)
	x, ok := podList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.PodList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text:        l.Items[i].Name,
			Description: string(l.Items[i].Status.Phase),
		}
	}
	sort.Sort(s)
	return s
}

func getPod(namespace, podName string) (corev1.Pod, bool) {
	x, ok := podList.Load(namespace)
	if !ok {
		return corev1.Pod{}, false
	}
	l, ok := x.(*corev1.PodList)
	if !ok || len(l.Items) == 0 {
		return corev1.Pod{}, false
	}
	for i := range l.Items {
		if podName == l.Items[i].Name {
			return l.Items[i], true
		}
	}
	return corev1.Pod{}, false
}

func getPortsFromPodName(namespace string, podName string) []prompt.Suggest {
	pod, found := getPod(namespace, podName)
	if !found {
		return []prompt.Suggest{}
	}

	// Extract unique ports
	portSet := make(map[int32]struct{})
	for i := range pod.Spec.Containers {
		ports := pod.Spec.Containers[i].Ports
		for j := range ports {
			portSet[ports[j].ContainerPort] = struct{}{}
		}
	}

	// Sort
	var ports []int
	for k := range portSet {
		ports = append(ports, int(k))
	}

	// Prepare suggestions
	s := make(Suggests, len(portSet))
	for i := range ports {
		s = append(s, prompt.Suggest{
			Text: fmt.Sprintf("%d:%d", ports[i], ports[i]),
		})
	}
	sort.Sort(s)
	return s
}

func getContainerNamesFromCachedPods(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchPods(client, namespace)

	x, ok := podList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.PodList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	// container name -> pod name
	set := make(map[string]string, len(l.Items))
	for i := range l.Items {
		for j := range l.Items[i].Spec.Containers {
			set[l.Items[i].Spec.Containers[j].Name] = l.Items[i].Name
		}
	}
	s := make(Suggests, len(l.Items))
	for key := range set {
		s = append(s, prompt.Suggest{
			Text:        key,
			Description: "Pod Name: " + set[key],
		})
	}
	sort.Sort(s)
	return s
}

func getContainerName(client *kubernetes.Clientset, namespace string, podName string) []prompt.Suggest {
	go fetchPods(client, namespace)

	pod, found := getPod(namespace, podName)
	if !found {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(pod.Spec.Containers))
	for i := range pod.Spec.Containers {
		s[i] = prompt.Suggest{
			Text:        pod.Spec.Containers[i].Name,
			Description: "",
		}
	}
	sort.Sort(s)
	return s
}

/* Daemon Sets */

var (
	daemonSetList *sync.Map
)

func fetchDaemonSetList(client *kubernetes.Clientset, namespace string) {
	key := "daemon_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.AppsV1().DaemonSets(namespace).List(metav1.ListOptions{})
	daemonSetList.Store(namespace, l)
	return
}

func getDaemonSetSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchDaemonSetList(client, namespace)
	x, ok := daemonSetList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(appsv1.DaemonSetList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Deployment */

var (
	deploymentList *sync.Map
)

func fetchDeployments(client *kubernetes.Clientset, namespace string) {
	key := "deployment_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.AppsV1().Deployments(namespace).List(metav1.ListOptions{})
	deploymentList.Store(namespace, l)
	return
}

func getDeploymentSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchDeployments(client, namespace)
	x, ok := deploymentList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*appsv1.DeploymentList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Endpoint */

var (
	endpointList *sync.Map
)

func fetchEndpoints(client *kubernetes.Clientset, namespace string) {
	key := "endpoint_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
	endpointList.Store(namespace, l)
	return
}

func getEndpointsSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchEndpoints(client, namespace)
	x, ok := endpointList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.EndpointsList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Events */

var (
	eventList *sync.Map
)

func fetchEvents(client *kubernetes.Clientset, namespace string) {
	key := "event_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Events(namespace).List(metav1.ListOptions{})
	eventList.Store(namespace, l)
	return
}

func getEventsSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchEvents(client, namespace)
	x, ok := eventList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.EventList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Node */

var (
	nodeList atomic.Value
)

func fetchNodeList(client *kubernetes.Clientset) {
	key := "node"
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Nodes().List(metav1.ListOptions{})
	nodeList.Store(l)
	return
}

func getNodeSuggestions(client *kubernetes.Clientset) []prompt.Suggest {
	go fetchNodeList(client)
	l, ok := nodeList.Load().(*corev1.NodeList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Secret */

var (
	secretList *sync.Map
)

func fetchSecretList(client *kubernetes.Clientset, namespace string) {
	key := "secret_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Secrets(namespace).List(metav1.ListOptions{})
	secretList.Store(namespace, l)
	return
}

func getSecretSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchSecretList(client, namespace)
	x, ok := secretList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.SecretList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Ingress */

var (
	ingressList *sync.Map
)

func fetchIngresses(client *kubernetes.Clientset, namespace string) {
	key := "ingress_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.ExtensionsV1beta1().Ingresses(namespace).List(metav1.ListOptions{})
	ingressList.Store(namespace, l)
}

func getIngressSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchIngresses(client, namespace)

	x, ok := ingressList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*extensionsv1beta1.IngressList)
	if !ok {
		debug.Log("must not reach here")
		return []prompt.Suggest{}
	}
	if len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* LimitRange */

var (
	limitRangeList *sync.Map
)

func fetchLimitRangeList(client *kubernetes.Clientset, namespace string) {
	key := "limit_range_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().LimitRanges(namespace).List(metav1.ListOptions{})
	limitRangeList.Store(namespace, l)
	return
}

func getLimitRangeSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchLimitRangeList(client, namespace)
	x, ok := limitRangeList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.NamespaceList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* NameSpaces */

func getNameSpaceSuggestions(namespaceList *corev1.NamespaceList) []prompt.Suggest {
	if namespaceList == nil || len(namespaceList.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(namespaceList.Items))
	for i := range namespaceList.Items {
		s[i] = prompt.Suggest{
			Text: namespaceList.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Persistent Volume Claims */

var (
	persistentVolumeClaimsList *sync.Map
)

func fetchPersistentVolumeClaimsList(client *kubernetes.Clientset, namespace string) {
	key := "persistent_volume_claims" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().PersistentVolumeClaims(namespace).List(metav1.ListOptions{})
	persistentVolumeClaimsList.Store(namespace, l)
	return
}

func getPersistentVolumeClaimSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchPersistentVolumeClaimsList(client, namespace)
	x, ok := persistentVolumeClaimsList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.PersistentVolumeClaimList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Persistent Volumes */

var (
	persistentVolumesList atomic.Value
)

func fetchPersistentVolumeList(client *kubernetes.Clientset) {
	key := "persistent_volume"
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	persistentVolumesList.Store(l)
	return
}

func getPersistentVolumeSuggestions(client *kubernetes.Clientset) []prompt.Suggest {
	go fetchPersistentVolumeList(client)
	l, ok := persistentVolumesList.Load().(*corev1.PersistentVolumeList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Pod Security Policies */

var (
	podSecurityPolicyList atomic.Value
)

func fetchPodSecurityPolicyList(client *kubernetes.Clientset) {
	key := "pod_security_policy"
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.ExtensionsV1beta1().PodSecurityPolicies().List(metav1.ListOptions{})
	podSecurityPolicyList.Store(l)
	return
}

func getPodSecurityPolicySuggestions(client *kubernetes.Clientset) []prompt.Suggest {
	go fetchPodSecurityPolicyList(client)
	l, ok := podSecurityPolicyList.Load().(policyv1beta1.PodSecurityPolicyList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Pod Templates */

var (
	podTemplateList *sync.Map
)

func fetchPodTemplateList(client *kubernetes.Clientset, namespace string) {
	key := "pod_template_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().PodTemplates(namespace).List(metav1.ListOptions{})
	podTemplateList.Store(namespace, l)
	return
}

func getPodTemplateSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchPodTemplateList(client, namespace)
	x, ok := podTemplateList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.PodTemplateList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Replica Sets */

var (
	replicaSetList *sync.Map
)

func fetchReplicaSetList(client *kubernetes.Clientset, namespace string) {
	key := "replica_set_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.AppsV1beta2().ReplicaSets(namespace).List(metav1.ListOptions{})
	replicaSetList.Store(namespace, l)
	return
}

func getReplicaSetSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchReplicaSetList(client, namespace)
	x, ok := replicaSetList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(appsv1.ReplicaSetList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Replication Controller */

var (
	replicationControllerList *sync.Map
)

func fetchReplicationControllerList(client *kubernetes.Clientset, namespace string) {
	key := "replication_controller" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().ReplicationControllers(namespace).List(metav1.ListOptions{})
	replicationControllerList.Store(namespace, l)
	return
}

func getReplicationControllerSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchReplicationControllerList(client, namespace)
	x, ok := replicationControllerList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.ReplicationControllerList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Resource quotas */

var (
	resourceQuotaList *sync.Map
)

func fetchResourceQuotaList(client *kubernetes.Clientset, namespace string) {
	key := "resource_quota" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().ResourceQuotas(namespace).List(metav1.ListOptions{})
	resourceQuotaList.Store(namespace, l)
	return
}

func getResourceQuotasSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchResourceQuotaList(client, namespace)
	x, ok := resourceQuotaList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.ResourceQuotaList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Service Account */

var (
	serviceAccountList *sync.Map
)

func fetchServiceAccountList(client *kubernetes.Clientset, namespace string) {
	key := "service_account_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().ServiceAccounts(namespace).List(metav1.ListOptions{})
	serviceAccountList.Store(namespace, l)
	return
}

func getServiceAccountSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchServiceAccountList(client, namespace)
	x, ok := serviceAccountList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.ServiceAccountList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Service */

var (
	serviceList *sync.Map
)

func fetchServiceList(client *kubernetes.Clientset, namespace string) {
	key := "service_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.CoreV1().Services(namespace).List(metav1.ListOptions{})
	serviceList.Store(namespace, l)
	return
}

func getServiceSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchServiceList(client, namespace)
	x, ok := serviceList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*corev1.ServiceList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text: l.Items[i].Name,
		}
	}
	sort.Sort(s)
	return s
}

/* Job */

var (
	jobList *sync.Map
)

func fetchJobs(client *kubernetes.Clientset, namespace string) {
	key := "job_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.BatchV1().Jobs(namespace).List(metav1.ListOptions{})
	jobList.Store(namespace, l)
}

func getJobSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchJobs(client, namespace)
	x, ok := jobList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*batchv1.JobList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		s[i] = prompt.Suggest{
			Text:        l.Items[i].Name,
			Description: l.Items[i].Status.StartTime.String(),
		}
	}
	sort.Sort(s)
	return s
}

/* Cronjob */

var (
	cronJobList *sync.Map
)

func fetchCronJobs(client *kubernetes.Clientset, namespace string) {
	key := "cronjob_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.BatchV1beta1().CronJobs(namespace).List(metav1.ListOptions{})
	cronJobList.Store(namespace, l)
}

func getCronJobSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchCronJobs(client, namespace)
	x, ok := cronJobList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*batchv1beta1.CronJobList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		desc := "last execute at: "
		if l.Items[i].Status.LastScheduleTime != nil {
			desc += l.Items[i].Status.LastScheduleTime.Time.String()
		} else {
			desc = "not execute yet"
		}

		s[i] = prompt.Suggest{
			Text:        l.Items[i].Name,
			Description: desc,
		}
	}
	sort.Sort(s)
	return s
}

/* hpa */

var (
	hpaList *sync.Map
)

func fetchHpas(client *kubernetes.Clientset, namespace string) {
	key := "hpa_" + namespace
	if !shouldFetch(key) {
		return
	}
	updateLastFetchedAt(key)

	l, _ := client.AutoscalingV2beta2().HorizontalPodAutoscalers(namespace).List(metav1.ListOptions{})
	hpaList.Store(namespace, l)
}

func getHpaSuggestions(client *kubernetes.Clientset, namespace string) []prompt.Suggest {
	go fetchHpas(client, namespace)
	x, ok := hpaList.Load(namespace)
	if !ok {
		return []prompt.Suggest{}
	}
	l, ok := x.(*v2beta2.HorizontalPodAutoscalerList)
	if !ok || len(l.Items) == 0 {
		return []prompt.Suggest{}
	}
	s := make(Suggests, len(l.Items))
	for i := range l.Items {
		desc := "last scaled at: "
		if l.Items[i].Status.LastScaleTime != nil {
			desc += l.Items[i].Status.LastScaleTime.Time.String()
		} else {
			desc = "not scaled yet"
		}

		s[i] = prompt.Suggest{
			Text:        l.Items[i].Name,
			Description: desc,
		}
	}
	sort.Sort(s)
	return s
}
