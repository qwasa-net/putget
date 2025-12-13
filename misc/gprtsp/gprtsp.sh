#!/bin/bash
# set -x

for cmd in hexdump ffmpeg curl openssl hexdump; do
    if ! command -v "$cmd" >/dev/null 2>&1; then
        echo "error: required command '$cmd' not found -- exiting" >&2
        exit 1
    fi
done

SLEEP_GETPUT=99
TIMEOUTS=30
FFMPEG_OPTS="-vf select=eq(n\\,11)"
CSEK=$(tr -dc 'A-Za-z0-9' < /dev/urandom 2>/dev/null | head -c 24)
SSEK=""

CONTENTTYPE="image/jpeg"
PUTGETHEADER="X-GP-CLIENT: gprtsp"

TS_LABEL_FORMAT="%Y-%m-%d %H:%M:%S"
TS_LABEL_COLOR="yellow"
TS_LABEL_BGCOLOR="black"
TS_LABEL_SIZE=30

# load env file
if [ -n "$1" ]; then
    source "$1"
fi

HKEY=$(echo -n "$CSEK" | openssl dgst -sha256 -binary | hexdump -v -e '/1 "%02x"' | tr -d ' \n')

if [ -n "$SSEK" ]; then
    PUTGETHEADER="X-SSE-C: ${SSEK}"
else
    PUTGETHEADER="X-SSE-C:"
fi

capture_image() {
    local url="$1"
    local output="$2"

    timeout ${TIMEOUTS} \
    ffmpeg -loglevel fatal -y \
    -i "${url}" \
    $FFMPEG_OPTS -an -f image2 -frames:v 1 -update true \
    "${output}"
}

annotate_image() {
    local image="$1"
    if [ -n "${TS_LABEL_FORMAT}" ]; then
        DATE_STR=$(date +"${TS_LABEL_FORMAT}")
        mogrify \
        -gravity SouthWest \
        -pointsize "${TS_LABEL_SIZE}" \
        -fill "${TS_LABEL_COLOR}" \
        -undercolor "${TS_LABEL_BGCOLOR}" \
        -annotate +10+10 "${DATE_STR}" "${image}"
    fi
}

encrypt_image() {
    local input="$1"
    local output="$2"
    local key="$3"

    local iv
    iv=$(dd if=/dev/urandom bs=16 count=1 2>/dev/null | base64 | head -c 16)
    local iv_hex
    iv_hex=$(echo -n "$iv" | hexdump -v -e '/1 "%02x"')

    openssl enc -aes-256-ctr -K "${key}" -iv "${iv_hex}" -in "${input}" -out "${input}.aes"
    echo -n "$iv" > "${output}"
    cat "${input}.aes" >> "${output}"
}

upload_image() {
    local file="$1"
    local url="$2"

    timeout ${TIMEOUTS} \
    curl $CURLOPTIONS \
    -X POST "${url}" \
    --user "${PUTGETAUTH}" \
    -H "Host: ${PUTGETHOST}" \
    -H "Content-type: ${CONTENTTYPE}" \
    -H "${PUTGETHEADER}" \
    --data-binary @"${file}"
}

while true
do
    sleep $SLEEP_GETPUT
    for STREAMURL in $STREAMURLS
    do
        C=1

        capture_image "${STREAMURL}" "${IMAGEFILE}"

        if [ ! -s "$IMAGEFILE" ]; then
            continue
        fi

        annotate_image "${IMAGEFILE}"

        encrypt_image "${IMAGEFILE}" "${IMAGEFILE}.enc" "${HKEY}"

        upload_image "${IMAGEFILE}.enc" "${PUTGETURL}-${C}"

        C=$((C + 1))

    done
done
