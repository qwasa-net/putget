#!/bin/sh

export PATH=/system/sdcard/bin:/system/bin:/bin:/sbin:/usr/bin:/usr/sbin
export OPENSSL_CONF=/dev/null

export XXD="/bin/busybox xxd -p -c 256"
export BASE64="/bin/busybox base64"

source /root/getput_env.sh

HKEY=$(echo -n "${CSEK}" | openssl dgst -sha256 -binary | ${XXD} | tr -d ' \n')

if [ -n "$SSEK" ]; then
    PUTGETHEADER="X-SSE-C: ${SSEK}"
else
    PUTGETHEADER="X-SSE-C:"
fi

while true
do
    sleep $SLEEP_GETPUT
    timeout -t 60 curl "$CAMURL" --user "$CAMAUTH" --insecure --silent -o "$IMAGEFILE"
    if [ -s "$IMAGEFILE" ]
    then

        IV=$(dd if=/dev/urandom bs=16 count=1 2>/dev/null | ${BASE64} | head -c 16) # IV=$(openssl rand 16)
        IV_HEX=$(echo -n "$IV" | ${XXD} | tr -d ' \n')
        openssl enc -aes-256-ctr -K "$HKEY" -iv "$IV_HEX" -in "$IMAGEFILE" -out "$IMAGEFILE.aes"
        echo -n "$IV" > "$IMAGEFILE.enc"
        cat "$IMAGEFILE.aes" >> "$IMAGEFILE.enc"

        timeout -t 60 \
        curl $CURLOPTIONS --silent \
        -X POST "$PUTGETURL" \
        --user "$PUTGETAUTH" \
        -H "Host: $PUTGETHOST" \
        -H "Content-type: image/jpeg" \
        -H "$PUTGETHEADER" \
        --data-binary @"$IMAGEFILE.enc"

    fi
done
