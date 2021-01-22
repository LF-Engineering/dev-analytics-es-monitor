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
cd /root/go/src/github.com/LF-Engineering/dev-analytics-es-monitor || exit 3
git pull || exit 4
make || exit 5
repo="`cat repo_access.secret`"
if [ -z "$repo" ]
then
  echo "$0: missing repo_access.secret file"
  exit 6
fi
rm -rf dev-analytics-api
git clone "${repo}" || exit 7
cd dev-analytics-api || exit 8
git checkout "$1" || exit 9
cd .. || exit 10
function cleanup {
  rm -rf "${lock_file}" dev-analytics-api
}
> "${lock_file}"
trap cleanup EXIT
./esmonitor ./dev-analytics-api/app/services/lf/bootstrap/fixtures 2>&1 | tee -a run.log
