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

Save the configuration and then map the volume to registry:

```
docker run -p 5000:5000 -v $(pwd)/config.yml:/etc/docker/registry/config.yml registry:2
```

Launch argoos and be sure that "argoos.url" is accessible using: http://argoos.url:3000/healtz is responding.

Registry will now contact Argoos for each pushed layers and images.

# Configuration with Kubernetes

You may create service and replication-controller from misc/kuebernetes directory of the argoos project.

```bash
kubectl create -f https://raw.githubusercontent.com/Smile-SA/argoos/devel/misc/kubernetes/argoos-service.yml
kubectl create -f https://raw.githubusercontent.com/Smile-SA/argoos/devel/misc/kubernetes/argoos-rc.yml
```

Everything is configured to be append on "kube-system" namespace.

If you serve registry by kubernetes, you can add a configmap to mount on "`/etc/docker/registry/config.yml`" using the explanation given above. Change "`url`" to "argoos.kube-system" **without the port** (argoos kubernetes service listens on 80).

Argoos will contact "kubernetes" API that is generally "kubernetes.default" on "8080" port. You can change it using environment variables in argoos-rc.yml file.


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

