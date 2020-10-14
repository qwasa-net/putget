#!/bin/bash

FILES_ROOT="${FILES_ROOT:-/home/putget.qwasa.net/files}"
FILES_OWNER="${FILES_OWNER:-$USER}"
FILES_OLD="${FILES_OLD:-15}"
PUTGET_CONTAINER_NAME="${PUTGET_CONTAINER_NAME:-putget}"
DOCKER_BIN="${DOCKER_BIN:-/usr/bin/podman}"

#
echo "@ $0 `/usr/bin/date`"

# cleanup old files
/usr/bin/find "${FILES_ROOT}" -user "${FILES_OWNER}" -type f -mtime "+${FILES_OLD}" -delete

# rotate logs
/usr/sbin/logrotate --verbose --state /home/putget.qwasa.net/files/logrotate.state /home/putget.qwasa.net/deploy/logrotate.conf

# kill -HUP (NB: container will restrated by systemd)
"${DOCKER_BIN}" stop -t 1 "${PUTGET_CONTAINER_NAME}"

# anything is good
exit 0