package apiutils

import (
	"log"
	"testing"
	"time"
)

const testevent = `
{ "events": [
      {
         "id": "5a98bda7-df19-4fd7-887c-b56ff3209115",
         "timestamp": "2016-11-01T10:15:50.257780446Z",
         "action": "push",
         "target": {
            "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
            "size": 529,
            "digest": "sha256:9c7d6bef1e91ed762731d87542d26ccc9400ed8ba6dd4ad42bf1f2e7d0c0cdd0",
            "length": 529,
            "repository": "test/app",
            "url": "http://kube-master:5000/v2/centos/manifests/sha256:9c7d6bef1e91ed762731d87542d26ccc9400ed8ba6dd4ad42bf1f2e7d0c0cdd0",
            "tag": "latest"
         },
         "request": {
            "id": "28264817-c50f-4783-8470-e852f3b31019",
            "addr": "172.16.68.1:41654",
            "host": "kube-master:5000",
            "method": "PUT",
            "useragent": "docker/1.12.3 go/go1.6.3 git-commit/6b644ec kernel/4.8.4-200.fc24.x86_64 os/linux arch/amd64 UpstreamClient(Docker-Client/1.12.3 \\(linux\\))"
         },
         "actor": {},
         "source": {
            "addr": "e0966eaed542:5000",
            "instanceID": "3a9ae225-95fe-4f3e-9ff6-0fc8be2cf20d"
         }
      }
   ]
}
`

func TestNamespaces(t *testing.T) {
	ns := getNameSpaces()
	t.Log(ns)
}

func TestGetDeployments(t *testing.T) {
	ret := getDeployments()
	t.Log(ret)
}

func TestParseEvents(t *testing.T) {
	events := getEvents([]byte(testevent))
	log.Println(events)
}

func TestImpactedDeployments(t *testing.T) {
	go rollout()
	events := getEvents([]byte(testevent))
	for _, e := range events.Events {
		getImpactedDeployments(e)
	}
	time.Sleep(time.Second * 2)

}
