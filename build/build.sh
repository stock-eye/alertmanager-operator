#/bin/sh
set -x

CI_COMMIT_TAG=$(git describe --always --tags)

docker build -t linclaus/alertmanager-operator:latest -f build/package/Dockerfile .
docker push linclaus/alertmanager-operator:latest