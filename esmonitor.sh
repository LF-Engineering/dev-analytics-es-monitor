#!/bin/bash
date
if [ -z "$1" ]
then
  echo "$0: you need to specify environment/branch: test|prod"
  exit 1
fi
lock_file="/tmp/es_monitor_$1.lock"
if [ -f "${lock_file}" ]
then
  echo "$0: another esmonitor instance \"$1\" is still running, exiting"
  exit 2
fi
if [ -z "$MONITOR_DIR" ]
then
  MONITOR_DIR=/root/go/src/github.com/LF-Engineering/dev-analytics-es-monitor
fi
if [ -z "${ES_URL}" ]
then
  export ES_URL="`cat ./ES_URL.${1}.secret`"
fi
if [ -z "${ES_URL}" ]
then
  echo "$0: you need to specify ES_URL env variable"
  exit 3
fi
cd "$MONITOR_DIR" || exit 4
git pull || exit 5
make || exit 6
repo="`cat repo_access.secret`"
if [ -z "$repo" ]
then
  echo "$0: missing repo_access.secret file"
  exit 7
fi
rm -rf dev-analytics-api
git clone "${repo}" || exit 8
cd dev-analytics-api || exit 9
git checkout "$1" || exit 10
cd .. || exit 11
function cleanup {
  rm -rf "${lock_file}" dev-analytics-api
}
> "${lock_file}"
trap cleanup EXIT
./esmonitor ./dev-analytics-api/app/services/lf/bootstrap/fixtures 2>&1 | tee -a run.log
