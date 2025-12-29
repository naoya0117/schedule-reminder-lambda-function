#!/usr/bin/env bash
set -euo pipefail

localstack_ip="$(getent hosts localstack | awk '{print $1}')"
if [ -z "${localstack_ip}" ]; then
  echo "Failed to resolve localstack IP via getent." >&2
  exit 1
fi

sam local invoke ScheduleReminderFunction \
  --config-env local \
  --add-host "localstack:${localstack_ip}"
