package kube

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"time"
)

var (
	contextLastFetchedAt time.Time
	currentContext       string
	currentNamespace     string
)

/**
自定义的prefix生成器，主要是展示了context和ns

如果更换了context，需要重新reload一下client
*/
func (c *Completer) GetPs1() (ps1 string, useLivePrefix bool) {
	context, namespace := getCurrentContextInfo(c)
	return fmt.Sprintf("(%s|%s)> ", context, namespace), true
}

func getCurrentContextInfo(c *Completer) (context, namespace string) {
	if !(contextLastFetchedAt.IsZero() || time.Since(contextLastFetchedAt) > thresholdFetchInterval) {
		return currentContext, currentNamespace
	}

	return ReloadContextInfo(c)
}

func ReloadContextInfo(c *Completer) (context, namespace string) {
	contextLastFetchedAt = time.Now()

	cfg, err := clientcmd.NewDefaultClientConfigLoadingRules().Load()
	if err != nil {
		panic(err)
	}

	contexts := cfg.Contexts

	if currentContext != cfg.CurrentContext {
		go c.ReloadContext()
	}

	currentContext = cfg.CurrentContext
	currentNamespace = contexts[currentContext].Namespace
	return currentContext, currentNamespace
}
