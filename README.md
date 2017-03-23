# Argoos, push to registry to deploy on Kubernetes

Argoos is a service able to listen Docker Registry events, fetch corresponding Kubernetes deployment and ask for rollout to them. It takes advantage from Docker Registry "notifications" you should configure.

Deployment must indicate a label to provides rules to Argoos. Without that label, Argoos will never touch deployment.

For example:

```yaml
kind: Deployment
metadata:
    name: myapp
  labels:
    argoos.io/policy: all
```

Policies can be:

- "patch" : to deploy images only if "patch" version has changed
- "minor" : to deploy images only if "minor" version has changed
- "major" : to deploy images only if "major" version has changed
- "all" : to deploy whatever the version is set, even if the docker image release name doesn't respect "semver" (semantic version)

# How to configure registry

You docker registry must be configure to call webhook via "notification". The easiest method is to get the default config file like this:

```bash
docker run --rm --entrypoint=cat registry:2 /etc/docker/registry/config.yml | tee config.yml
version: 0.1
log:
  fields:
    service: registry
storage:
  cache:
    blobdescriptor: inmemory
  filesystem:
    rootdirectory: /var/lib/registry
http:
  addr: :5000
  headers:
    X-Content-Type-Options: [nosniff]
health:
  storagedriver:
    enabled: true
    interval: 10s
    threshold: 3
```

Now, edit "config.yml" to append notifications:

```yaml
notifications:
  endpoints:
    - name: argoos
      url: http://argoos.url:3000/event
      timeout: 500ms
      threshold: 5
      backoff: 1s
```

**Important !** If you deployed registry in kubernetes, and/or if your deployment uses other registry URL that the given in "push" command, Argoos will not be able to identify which hostname to test. So, to override used URL by Argoos, you may add that header:


```yaml
notifications:
  endpoints:
    - name: argoos
      url: http://argoos.url:3000/event
      timeout: 500ms
      threshold: 5
      backoff: 1s
      headers:
        X-Argoos-Registry-Name: ["host-to-use"]
```

Replace "host-to-use" by the used one in Kubernetes deployment. For example: "localhost:5000" if you bind registry port to the 5000 node port.

Save the configuration and then map the volume to registry:

```
docker run -p 5000:5000 -v $(pwd)/config.yml:/etc/docker/registry/config.yml registry:2
```

Launch argoos and be sure that "argoos.url" is accessible using: http://argoos.url:3000/healtz is responding.

Registry will now contact Argoos for each pushed layers and images.

# Configuration with Kubernetes

This section provides a basic way to add argoos and a docker registry. Note that the registry is not protected and you will probably need to adapt configuration to your requirements.

## Add Argoos in kube-system namespace

You may create service and replication-controller from misc/argoos directory of the argoos project.

```bash
# To ease api requests, you should create a serviceaccount
# that will be used by argoos
kubectl create serviceaccount argoos --namespace=kube-system
_ARGOOS_GIT="https://raw.githubusercontent.com/Smile-SA/argoos/master/misc"
kubectl create -f $_ARGOOS_GIT/argoos/argoos.svc.yml
kubectl create -f $_ARGOOS_GIT/argoos/argoos.deploy.yml
```

Everything is configured to be installed in "kube-system" namespace.

Now, there is Argoos running, we may configure a registry.

## Add registry in Kubernetes

A simple way to be able to work with Argoos is to add a registry in Kubernetes.

We provide a basic registry configuration to help. That configuration uses "argoos" url that is served by argoos service we've juste deployed.

```
_ARGOOS_GIT="https://raw.githubusercontent.com/Smile-SA/argoos/master/misc"
kubectl create -f $_ARGOOS_GIT/registry/registry.pvc.yaml
kubectl create -f $_ARGOOS_GIT/registry/registry.configmap.yaml
kubectl create -f $_ARGOOS_GIT/registry/registry.svc.yaml
kubectl create -f $_ARGOOS_GIT/registry/registry.deploy.yaml
kubectl create -f $_ARGOOS_GIT/registry/registry.ds.yaml
```

You must have a configured PersistentVolume. If not, clone our repository and change the deployment to remove/change "registry-data" volume. **Keep configmap volume !**

The provided configmap set notification header to replace the registry ip to "localhost". Argoos will be able to use "localhost" because we've installed a deamonset to contact registry on each node.

![](misc/registry-diagram.png)

You may now be able to push images from outside to node-ip:5000 and be able to set image url in deployment with "localhost:5000". See: https://github.com/kubernetes/kubernetes/tree/master/cluster/addons/registry page that explains what it does. Also, note that we've adapted daemonset in our repository to change app name.

# Restrict access

You may restrict access by asking X-Argoos-Token header from Docker registry.

```yaml
notifications:
  endpoints:
    - name: argoos
      url: http://argoos.url:3000/event
      headers:
          X-Argoos-Token: <you token here>
      timeout: 500ms
      threshold: 5
      backoff: 1s
```

By setting TOKEN environment variable or "-token" argument to the argoos command line, then "/event" api will respond "Unauthorized" if the token is not set and/or doesn't correspond.

It's **strongly recommended to set that token in a kubernetes "secret"**


```bash
$ TOKEN=$(tr -dc "[:alnum:]" < /dev/urandom | head -c32)
$ echo $TOKEN
$ kubectl create secret generic argoos-token --from-litteral=token=$TOKEN
```

Use the token in you registry configuration, then use "argoos-token" secret in argoos deployment:

```yaml
# ...
    - image: smilelab/argoos
      name: argoos
      env:
      - name: TOKEN
        valueFrom:
          secretKeyRef:
            name: argoos-token
            key: token
```

