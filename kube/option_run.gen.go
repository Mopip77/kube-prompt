// Code generated by 'option-gen'. DO NOT EDIT.

package kube

import (
	prompt "github.com/c-bata/go-prompt"
)

var runOptions = []prompt.Suggest{
	prompt.Suggest{Text: "--allow-missing-template-keys", Description: "If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats."},
	prompt.Suggest{Text: "--attach", Description: "If true, wait for the Pod to start running, and then attach to the Pod as if 'kubectl attach ...' were called.  Default false, unless '-i/--stdin' is set, in which case the default is true. With '--restart=Never' the exit code of the container process is returned."},
	prompt.Suggest{Text: "--command", Description: "If true and extra arguments are present, use them as the 'command' field in the container, rather than the 'args' field which is the default."},
	prompt.Suggest{Text: "--dry-run", Description: "If true, only print the object that would be sent, without sending it."},
	prompt.Suggest{Text: "--env", Description: "Environment variables to set in the container"},
	prompt.Suggest{Text: "--expose", Description: "If true, a public, external service is created for the container(s) which are run"},
	prompt.Suggest{Text: "--generator", Description: "The name of the API generator to use, see http://kubernetes.io/docs/user-guide/kubectl-conventions/#generators for a list."},
	prompt.Suggest{Text: "--hostport", Description: "The host port mapping for the container port. To demonstrate a single-machine container."},
	prompt.Suggest{Text: "--image", Description: "The image for the container to run."},
	prompt.Suggest{Text: "--image-pull-policy", Description: "The image pull policy for the container. If left empty, this value will not be specified by the client and defaulted by the server"},
	prompt.Suggest{Text: "--include-extended-apis", Description: "If true, include definitions of new APIs via calls to the API server. [default true]"},
	prompt.Suggest{Text: "-l", Description: "Comma separated labels to apply to the pod(s). Will override previous values."},
	prompt.Suggest{Text: "--labels", Description: "Comma separated labels to apply to the pod(s). Will override previous values."},
	prompt.Suggest{Text: "--leave-stdin-open", Description: "If the pod is started in interactive mode or with stdin, leave stdin open after the first attach completes. By default, stdin will be closed after the first attach completes."},
	prompt.Suggest{Text: "--limits", Description: "The resource requirement limits for this container.  For example, 'cpu=200m,memory=512Mi'.  Note that server side components may assign limits depending on the server configuration, such as limit ranges."},
	prompt.Suggest{Text: "--no-headers", Description: "When using the default or custom-column output format, don't print headers (default print headers)."},
	prompt.Suggest{Text: "-o", Description: "Output format. One of: json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=... See custom columns [http://kubernetes.io/docs/user-guide/kubectl-overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [http://kubernetes.io/docs/user-guide/jsonpath]."},
	prompt.Suggest{Text: "--output", Description: "Output format. One of: json|yaml|wide|name|custom-columns=...|custom-columns-file=...|go-template=...|go-template-file=...|jsonpath=...|jsonpath-file=... See custom columns [http://kubernetes.io/docs/user-guide/kubectl-overview/#custom-columns], golang template [http://golang.org/pkg/text/template/#pkg-overview] and jsonpath template [http://kubernetes.io/docs/user-guide/jsonpath]."},
	prompt.Suggest{Text: "--overrides", Description: "An inline JSON override for the generated object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field."},
	prompt.Suggest{Text: "--pod-running-timeout", Description: "The length of time (like 5s, 2m, or 3h, higher than zero) to wait until at least one pod is running"},
	prompt.Suggest{Text: "--port", Description: "The port that this container exposes.  If --expose is true, this is also the port used by the service that is created."},
	prompt.Suggest{Text: "--quiet", Description: "If true, suppress prompt messages."},
	prompt.Suggest{Text: "--record", Description: "Record current kubectl command in the resource annotation. If set to false, do not record the command. If set to true, record the command. If not set, default to updating the existing annotation value only if one already exists."},
	prompt.Suggest{Text: "-r", Description: "Number of replicas to create for this container. Default is 1."},
	prompt.Suggest{Text: "--replicas", Description: "Number of replicas to create for this container. Default is 1."},
	prompt.Suggest{Text: "--requests", Description: "The resource requirement requests for this container.  For example, 'cpu=100m,memory=256Mi'.  Note that server side components may assign requests depending on the server configuration, such as limit ranges."},
	prompt.Suggest{Text: "--restart", Description: "The restart policy for this Pod.  Legal values [Always, OnFailure, Never].  If set to 'Always' a deployment is created, if set to 'OnFailure' a job is created, if set to 'Never', a regular pod is created. For the latter two --replicas must be 1.  Default 'Always', for CronJobs `Never`."},
	prompt.Suggest{Text: "--rm", Description: "If true, delete resources created in this command for attached containers."},
	prompt.Suggest{Text: "--save-config", Description: "If true, the configuration of current object will be saved in its annotation. Otherwise, the annotation will be unchanged. This flag is useful when you want to perform kubectl apply on this object in the future."},
	prompt.Suggest{Text: "--schedule", Description: "A schedule in the Cron format the job should be run with."},
	prompt.Suggest{Text: "--service-generator", Description: "The name of the generator to use for creating a service.  Only used if --expose is true"},
	prompt.Suggest{Text: "--service-overrides", Description: "An inline JSON override for the generated service object. If this is non-empty, it is used to override the generated object. Requires that the object supply a valid apiVersion field.  Only used if --expose is true."},
	prompt.Suggest{Text: "--serviceaccount", Description: "Service account to set in the pod spec"},
	prompt.Suggest{Text: "-a", Description: "When printing, show all resources (default show all pods including terminated one.)"},
	prompt.Suggest{Text: "--show-all", Description: "When printing, show all resources (default show all pods including terminated one.)"},
	prompt.Suggest{Text: "--show-labels", Description: "When printing, show all labels as the last column (default hide labels column)"},
	prompt.Suggest{Text: "--sort-by", Description: "If non-empty, sort list types using this field specification.  The field specification is expressed as a JSONPath expression (e.g. '{.metadata.name}'). The field in the API resource specified by this JSONPath expression must be an integer or a string."},
	prompt.Suggest{Text: "-i", Description: "Keep stdin open on the container(s) in the pod, even if nothing is attached."},
	prompt.Suggest{Text: "--stdin", Description: "Keep stdin open on the container(s) in the pod, even if nothing is attached."},
	prompt.Suggest{Text: "--template", Description: "Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview]."},
	prompt.Suggest{Text: "-t", Description: "Allocated a TTY for each container in the pod."},
	prompt.Suggest{Text: "--tty", Description: "Allocated a TTY for each container in the pod."},
}
