#!/bin/bash

generate_sam_options() {
  sam_options=()
  [ -n "$CONTAINER_HOST" ] && sam_options+=("--container-host" "$CONTAINER_HOST")
  [ -n "$EXTRA_CONTAINER_SERVICE" ] &&
    sam_options+=("--add-host" "$EXTRA_CONTAINER_SERVICE:$(getent hosts "$EXTRA_CONTAINER_SERVICE" | awk '{print $1}')")
  [ "${DEBUG_MODE:-0}" = 1 ] && sam_options+=("--debug-port 9229")
  [ "${DEBUG:-0}" = 1 ] && sam_options+=("--debug")
}

# CMDが"sam local"で始まる場合、サーバー起動オプションを生成
if [[ "$1" == "sam" && "$2" == "local" ]]; then
  generate_sam_options
  exec bash -c "$* ${sam_options[*]}"
else
  exec bash -c "$@"
fi