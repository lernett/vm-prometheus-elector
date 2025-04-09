## vm-prometheus-elector
This  is a fork of [prometheus-elector](https://github.com/jlevesy/prometheus-elector). Сompared to the original project, it adds the ability to work with vmagent  and prometheus as part of victoria-metrics-operator and prometheus-operator respectively. It does this by simplifying the logic: instead of dealing with merging leader/follower configs, it just removes (for the follower) a config part responsible for scraping targets, which is identical for both vmagent and prometheus, but could also easily be customized if needed.

This is an example of how you can use it with vmagent.
```yaml
vmagent:
  spec:
    configReloaderExtraArgs:
      config-envsubst-file: "/etc/vmagent/config_out/vmagent_config.yaml"
      
    initContainers:
      - name: init-vmagent-elector
        image: lernett/vmagent-elector:1.0.1
        args:
          - -config=/etc/vmagent/config_out/vmagent_config.yaml
          - -output=/etc/vmagent/config_out/vmagent.env.yaml
          - -init
        command:
          - /app/elector-cmd
        securityContext:
          capabilities:
            drop: [ "ALL" ]
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
        volumeMounts:
          - mountPath: /etc/vmagent/config_out
            name: config-out
          - mountPath: /etc/vmagent/config
            name: config


    containers:
      - name: vmagent-elector
        image: lernett/vmagent-elector:1.0.1
        args:
          - -lease-name=vm-elector-lease
          - -lease-namespace=monitoring
          - -config=/etc/vmagent/config_out/vmagent_config.yaml
          - -output=/etc/vmagent/config_out/vmagent.env.yaml
          - -notify-http-url=http://127.0.0.1:8429/-/reload
          - -readiness-http-url=http://127.0.0.1:8429/health
          - -healthcheck-http-url=http://127.0.0.1:8429/health
          - -api-listen-address=:9095
        command:
          - /app/elector-cmd
        ports:
          - name: http-elector
            containerPort: 9095
            protocol: TCP
        securityContext:
          capabilities:
            drop: [ "ALL" ]
          readOnlyRootFilesystem: true
          runAsNonRoot: true
          runAsUser: 1000
        volumeMounts:
          - mountPath: /etc/vmagent/config_out
            name: config-out
          - mountPath: /etc/vmagent/config
            name: config
```
Note that you also need to grant permissions to access lease resources, you can use something like this:
```yaml
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: vm-elector-role
  namespace: monitoring
rules:
  - apiGroups:
      - coordination.k8s.io
    resources:
      - leases
    verbs:
      - get
      - list
      - watch
      - create
      - update
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: vm-elector-rolebinding
  namespace: monitoring
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: vm-elector-role
subjects:
  - kind: ServiceAccount
    name: vmagent-vmagent-victoria-metrics-k8s-stack
    namespace: monitoring
```
