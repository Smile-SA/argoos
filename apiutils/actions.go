package apiutils

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var (
	KubeMasterURL       = "http://kube-master:8080"
	SkipSSLVerification = true
	Updates             = map[string]*Update{}
	stopRollout         = make(chan int)
	rolloutStarted      = false
)

const (
	deploymentApiVersion = "extensions/v1beta1"
	argoosLabel          = "argoos.io/policy"
)

func init() {
	Updates = make(map[string]*Update)
}

// Avoid container duplication - sometimes regitry sends
// two events, one for the tag version +  one for the "latest" tag.
// We should only use tag version in that case.
func cleanupContainerDupli(containers []Container) []Container {
	names := map[string]Container{}
	for _, c := range containers {
		if _, ok := names[c.Name]; ok {
			p := strings.Split(c.Image, ":")
			version := p[len(p)-1]
			if version != "lastest" {
				names[c.Name] = c
			}
		} else {
			names[c.Name] = c
		}
	}
	ctn := []Container{}
	for _, c := range names {
		ctn = append(ctn, c)
	}
	return ctn
}

// Check Updates map to send new deployment configuration to Kubernetes.
func rollout() {
	rolloutStarted = true
	for {
		select {
		case <-stopRollout:
			return
		case <-time.Tick(1 * time.Second):
			for api, update := range Updates {
				// cleanup
				update.Containers = cleanupContainerDupli(update.Containers)

				log.Println("updating...", api, update)

				// initialize a configuration update map that will be
				// encoded in json
				data := map[string]interface{}{
					"spec": map[string]interface{}{
						"template": map[string]interface{}{
							"spec": update,
						},
					},
				}
				c, _ := json.Marshal(data)
				buff := bytes.NewReader(c)
				// Create a request to Kubernetes api as PATCH
				req, err := http.NewRequest(http.MethodPatch, api, buff)
				if err != nil {
					log.Println(err)
				}
				req.Header.Add("Content-Type", "application/merge-patch+json")

				// Avoid SSL veification if needed
				client := http.Client{}
				if SkipSSLVerification {
					client.Transport = &http.Transport{
						TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
					}
				}
				// Launch deployment update
				resp, err := client.Do(req)
				rb, _ := ioutil.ReadAll(resp.Body)
				if err != nil {
					log.Println(req)
					log.Println(resp)
					log.Println(err)
					log.Println(string(rb))
				} else {
					log.Println(req.URL.String(), "called", resp.StatusCode, resp.Status)
					log.Println(string(rb))
				}

				// and remove update from the list
				delete(Updates, api)
			}
		}
	}

}

// append update to the global Updates var that is parsed by rollout() function.
func updateDeployment(dep map[string]string, event Event) {
	// prepare map key to be the URL to hit
	api := fmt.Sprintf("%s/apis/%s/namespaces/%s/deployments/%s",
		deploymentApiVersion,
		KubeMasterURL,
		dep["namespace"],
		dep["name"])

	if _, ok := Updates[api]; !ok {
		// prepare update in map
		Updates[api] = &Update{
			Containers: make([]Container, 0),
		}
	}

	containers := Updates[api].Containers
	containers = append(containers, Container{
		Image: fmt.Sprintf("%s/%s:%s",
			event.Request.Host,
			event.Target.Repository,
			event.Target.Tag),
		Name: dep["container"],
	})
	Updates[api].Containers = containers
}

// Fetch namespaces from kubernetes api.
func getNameSpaces() []string {
	api := "/api/v1/namespaces"
	resp, err := http.Get(KubeMasterURL + api)
	if err != nil {
		log.Println(err)
		return nil
	}
	decoder := json.NewDecoder(resp.Body)
	namespaces := APIResponse{}
	decoder.Decode(&namespaces)
	ret := []string{}
	for _, namespace := range namespaces.Items {
		ret = append(ret, namespace.Metadata.Name)
	}

	return ret
}

// fetch each deployment in all namespaces.
func getDeployments() []map[string]string {
	namespaces := getNameSpaces()
	ret := []map[string]string{}
	for _, ns := range namespaces {
		api := "/apis/%s/namespaces/%s/deployments"
		resp, _ := http.Get(KubeMasterURL + fmt.Sprintf(api, deploymentApiVersion, ns))
		spec := &APIResponse{}
		d := json.NewDecoder(resp.Body)
		d.Decode(spec)
		for _, item := range spec.Items {
			for _, t := range item.Spec.Template.Spec.Containers {
				policy := "none"
				if v, ok := item.Metadata.Labels[argoosLabel]; ok {
					switch v := v.(type) {
					case string:
						policy = v
					}
				}
				ret = append(ret, map[string]string{
					"namespace": ns,
					"name":      item.Metadata.Name,
					"container": t.Name,
					"image":     t.Image,
					"policy":    policy,
				})
			}
		}
	}
	return ret
}

// parse deployments and check policy label to know what to do.
func getImpactedDeployments(event Event) {
	deployments := getDeployments()
	eimage := fmt.Sprintf("%s/%s",
		event.Request.Host,
		event.Target.Repository)
	for _, d := range deployments {
		imgver := strings.Split(d["image"], ":")
		image := strings.Join(imgver[:len(imgver)-1], ":")
		if image != eimage {
			log.Println("Image has changed, no update for security reason")
			continue
		}

		update := false
		switch d["policy"] {
		case "none":
			// do nothing !
		case "latest":
			// if updated image is "latest" and policy is "latest", then update is ok
			if event.Target.Tag == "latest" {
				update = true
			}
		case "all":
			// upate for any image version
			update = true
		default:
			// compare image sent in registry and image found in deployment
			maj, min, patch := getVersion(event.Target.Tag)
			imaj, imin, ipatch := getVersion(imgver[len(imgver)-1])

			switch d["policy"] {
			case "patch":
				if maj == imaj && min == imin && patch > ipatch {
					update = true
				}
				fallthrough
			case "minor":
				if maj == imaj && min > imin {
					update = true
				}
				fallthrough
			case "major":
				if maj > imaj {
					update = true
				}
			}
		}

		if !update {
			log.Println("SKIPPED", d["namespace"], d["name"], d["image"], d["tag"])
			continue
		}
		updateDeployment(d, event)
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
