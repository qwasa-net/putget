#!/bin/sh

export PATH=/system/sdcard/bin:/system/bin:/bin:/sbin:/usr/bin:/usr/sbin
export OPENSSL_CONF=/dev/null

export XXD="/bin/busybox xxd"
# export XXD="xxd"

source /root/getput_env.sh

HKEY=$(echo -n "$CSEK" | openssl dgst -sha256 -binary | ${XXD} -p -c 256 | tr -d ' \n')

while true
do
    sleep $SLEEP_GETPUT
    timeout -t 60 curl "$CAMURL" --user "$CAMAUTH" --insecure --silent -o "$IMAGEFILE"
    if [ -s "$IMAGEFILE" ]
    then
        IV=$(openssl rand 16) #IV=$(tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 16)
        IV_HEX=$(echo -n "$IV" | ${XXD} -p -c 256 | tr -d ' \n')
        # echo openssl enc -aes-256-ctr -K "$HKEY" -iv "$IV_HEX" -in "$IMAGEFILE" -out "$IMAGEFILE.aes"
        openssl enc -aes-256-ctr -K "$HKEY" -iv "$IV_HEX" -in "$IMAGEFILE" -out "$IMAGEFILE.aes"
        echo -n "$IV" > "$IMAGEFILE.enc"
        cat "$IMAGEFILE.aes" >> "$IMAGEFILE.enc"

        timeout -t 60 \
        curl $CURLOPTIONS --insecure --silent \
        -X POST "$PUTGETURL" \
        --user "$PUTGETAUTH" \
        -H "Host: $PUTGETHOST" \
        -H "Content-type: image/jpeg" \
        -H "$PUTGETHEADER" \
        --data-binary @"$IMAGEFILE.enc"
    fi
done
