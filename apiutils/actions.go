package apiutils

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

var (
	// KubeMasterURL, URL to kubernetes master.
	KubeMasterURL = "http://kube-master:8080"
	// SkipSSLVerification allows to connect kubernetes without verifying certificates.
	SkipSSLVerification = true

	// CAFile to use with kubernetes if any.
	CAFile = ""

	// CertFile to use with kubernetes if any.
	CertFile = ""

	// KeyFile private key to use with kubernetes, if any.
	KeyFile = ""

	toUpdate       = make(chan *v1beta1.Deployment)
	stopRollout    = make(chan int)
	rolloutStarted = false
	kubeConfig     = &rest.Config{}
	kube           = &kubernetes.Clientset{}
	versionreg     = regexp.MustCompile(`:\d+\.\d+\.\d+$`)
	latestreg      = regexp.MustCompile(`:latest$`)
)

const (
	argoosLabel = "argoos.io/policy"
)

func init() {
	kubeConfig.Host = KubeMasterURL
	kube, _ = kubernetes.NewForConfig(kubeConfig)
}

// Check Updates map to send new deployment configuration to Kubernetes.
//
// TODO: deployments can have several container updates but we don't check this. Maybe
// the solution is to go back to a pool system or be sure that registry finished
// the entire push processes to launch deployment updates.
func rollout() {
	for {
		select {
		case <-stopRollout:
			return
		case u := <-toUpdate:
			go func(u *v1beta1.Deployment) {
				log.Println("Deploying", u)
				if _, err := kube.Deployments(u.Namespace).Update(u); err != nil {
					log.Println(err)
				}
			}(u)
		}
	}
}

// Fetch namespaces from kubernetes api.
func getNameSpaces() []string {
	ns := kube.Namespaces()
	ret := []string{}
	n, _ := ns.List(v1.ListOptions{})
	for _, n := range n.Items {
		ret = append(ret, n.GetName())
	}
	return ret
}

// fetch each deployment in all namespaces.
// REFACTO
func getDeployments() []*v1beta1.DeploymentList {
	ns := kube.Namespaces()
	n, _ := ns.List(v1.ListOptions{})
	ret := []*v1beta1.DeploymentList{}
	for _, n := range n.Items {
		dep := kube.Deployments(n.GetName())
		l, _ := dep.List(v1.ListOptions{})
		ret = append(ret, l)
	}
	return ret
}

func checkToUpdate(event Event, d v1beta1.Deployment, policy string) {
	pvMajor, pvMinor, pvPatch := getVersion(event.Target.Tag)
	for i, c := range d.Spec.Template.Spec.Containers {
		vMajor, vMinor, vPatch := getVersion(c.Image)
		update := false
		switch policy {
		case "all":
			update = true
		case "major":
			update = pvMajor > vMajor
			fallthrough
		case "minor":
			update = update || (pvMajor == vMajor && pvMinor > vMinor)
			fallthrough
		case "patch":
			update = update || (pvMajor == vMajor && pvMinor == vMinor && pvPatch > vPatch)
		}
		c.Image = fmt.Sprintf("%s/%s:%s", event.Request.Host, event.Target.Repository, event.Target.Tag)
		d.Spec.Template.Spec.Containers[i] = c
		if update {
			go func() {
				toUpdate <- &d
			}()
		}
	}
}

// parse deployments and check policy label to know what to do.
func getImpactedDeployments(event Event) {
	deployments := getDeployments()
	eimage := fmt.Sprintf("%s/%s", event.Request.Host, event.Target.Repository)
	for _, d := range deployments {
		for _, i := range d.Items {
			labels := i.GetLabels()
			policy := ""
			if v, ok := labels[argoosLabel]; ok {
				policy = v
			} else {
				continue
			}
			for _, c := range i.Spec.Template.Spec.Containers {
				// Remove version if any
				dimage := versionreg.ReplaceAllString(c.Image, "")
				dimage = latestreg.ReplaceAllString(dimage, "")
				if dimage == eimage {
					// there, pushed image corresponds to the deployment image
					// so we can check if we should update it
					checkToUpdate(event, i, policy)
				}
			}
		}
	}
}

// decode json data from event body.
func getEvents(c []byte) Events {

	events := Events{}
	reduced := []Event{}
	err := json.Unmarshal(c, &events)
	if err != nil {
		log.Println(err)
		return events
	}
	for _, event := range events.Events {
		// only get "push" events
		if event.Action == "push" && len(event.Target.Tag) > 0 {
			reduced = append(reduced, event)
		}
	}
	events.Events = reduced
	return events
}

// decompose version string in major, minor, patch list.
func getVersion(a string) (int, int, int) {
	v := strings.Split(a, ".")
	switch len(v) {
	case 0:
		v = append(v, "0")
		fallthrough
	case 1:
		v = append(v, "0")
		fallthrough
	case 2:
		v = append(v, "0")
	}
	version := []int{}
	for _, i := range v {
		s, _ := strconv.Atoi(i)
		version = append(version, s)
	}
	return version[0], version[1], version[2]
}

// GetEvents returns events from registry message
// given from webook body.
func GetEvents(c []byte) Events {
	return getEvents(c)
}

// ImpactedDeployments will fetch deployments using the
// repository image found in event to be impacted. It will check
// label to know if it should be entered in updates list that are
// managed by rollout goroutine.
func ImpactedDeployments(event Event) {
	getImpactedDeployments(event)
}

// StartRollout starts a goroutine on rollout() function
// that is a loop checking updates to send to Kubernetes Deployment
// objects.
func StartRollout() {
	go rollout()
}

// StopRollout stops rollout goroutine.
func StopRollout() {
	if rolloutStarted {
		stopRollout <- 1
	}
	rolloutStarted = false
}
