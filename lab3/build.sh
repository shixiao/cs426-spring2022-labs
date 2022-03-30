#!/bin/bash

set -e

image=$1
shift
files=$@

echo "Cleaning up old remote builds"
kubectl delete pod -l app=remote-build

build_id=$(($RANDOM % 10000))

tar czf $build_id.tar.gz $files

url=$(curl -fu $NETID:$PASSWORD -F "file=@$build_id.tar.gz" https://upload.cs426.cloud)
if [ $? -ne 0 ]; then
  echo "Failed to upload file"
  exit 1
fi
echo "Uploaded build to $url"

rm $build_id.tar.gz

internal_url=$(echo "$url" | sed -E "s,https://upload.cs426,https://upload-internal.cs426,")
echo "Using internal URL: $internal_url"

kaniko_pod_defn="
apiVersion: v1
kind: Pod
metadata:
  name: build-$build_id
  labels:
    app: remote-build
spec:
  containers:
  - name: kaniko
    image: gcr.io/kaniko-project/executor:latest
    args:
    - \"--dockerfile=Dockerfile\"
    - \"--context=$internal_url\"
    - \"--destination=$image\"
    resources:
      requests:
        memory: "2Gi"
        cpu: 100m
      limits:
        memory: "3Gi"
        cpu: 400m
    volumeMounts:
    - name: docker-config
      mountPath: /kaniko/.docker/
  restartPolicy: Never
  volumes:
  - name: docker-config
    configMap:
      name: docker-config 
"

echo "$kaniko_pod_defn" | kubectl create -f -

echo "Created pod: build-$build_id"
echo "Check on it with kubectl get pod build-$build_id"
echo "Once it is running, you can view logs with kubectl logs -f build-$build_id"
