package kube

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	contextLastFetchedAt time.Time
	currentContext string
	currentNamespace string
)

/**
自定义的prefix生成器，主要是展示了context和ns
*/
func GetPs1() (ps1 string, useLivePrefix bool) {
	context, namespace := getCurrentContextInfo()
	return fmt.Sprintf("(%s|%s)> ", context, namespace), true
}

func getCurrentContextInfo() (context, namespace string) {
	if !(contextLastFetchedAt.IsZero() || time.Since(contextLastFetchedAt) > thresholdFetchInterval) {
		return currentContext, currentNamespace
	}

	contextLastFetchedAt = time.Now()

	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		panic(err)
	}

	contexts := cfg.Contexts

	currentContext = cfg.CurrentContext
	currentNamespace = contexts[currentContext].Namespace
	return currentContext, currentNamespace
}
