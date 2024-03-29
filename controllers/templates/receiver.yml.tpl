{{- $name := get . "name"}}
{{- $namespace := get . "namespace"}}
{{- $ownerRefs := get . "ownerRefs"}}

{{- $tolerations := get . "tolerations" |default list }}
{{- $nodeSelector := get . "node-selector" |default dict }}
{{- $devInfo := get . "dev-info" | default ""}}

apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ $name }}
  annotations:
    kloudlite.io/account.name: {{ $name }}
  labels:
    kloudlite.io/wg-deployment: "true"
    kloudlite.io/wg-device.name: {{ $name }}
  ownerReferences: {{ $ownerRefs| toJson}}
  namespace: {{ $namespace }}
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      kloudlite.io/pod-type: wireguard-server
      kloudlite.io/device: {{$name}}
  template:
    metadata:
      labels:
        kloudlite.io/pod-type: wireguard-server
        kloudlite.io/device: {{$name}}
    spec:
      nodeSelector: {{$nodeSelector | toJson}}
      tolerations: {{$tolerations | toJson}}
      containers:
      - name: wireguard
        imagePullPolicy: IfNotPresent
        image: ghcr.io/linuxserver/wireguard
        securityContext:
          capabilities:
            add:
              - NET_ADMIN
              - SYS_MODULE
          privileged: true
        volumeMounts:
          - name: wg-config
            mountPath: /etc/wireguard/wg0.conf
            subPath: wg0.conf
          - name: host-volumes
            mountPath: /lib/modules
          - mountPath: /etc/sysctl.conf
            name: sysctl
            subPath: sysctl.conf
        ports:
        - containerPort: 51820
          protocol: UDP
        resources:
          requests:
            memory: 10Mi
            # cpu: "100m"
          limits:
            memory: "10Mi"
            # cpu: "200m"

      volumes:
        - name: sysctl
          secret:
            items:
            - key: sysctl
              path: sysctl.conf
            secretName: "wg-configs-{{$name}}"
        - name: wg-config
          secret:
            secretName: "wg-configs-{{$name}}"
            items:
              - key: server-config
                path: wg0.conf
        - name: host-volumes
          hostPath:
            path: /lib/modules
            type: Directory

      dnsPolicy: Default
      priorityClassName: system-cluster-critical
      restartPolicy: Always
      schedulerName: default-scheduler
      terminationGracePeriodSeconds: 30