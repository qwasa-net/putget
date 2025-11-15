#!/bin/sh
PIDFILE='/run/getput.pid'

status()
{  PID=$(cat $PIDFILE)
   if [ "$PID" != "" ]; then
   cat $PIDFILE
   fi
}

start()
{
  LOG=/dev/null
  echo "Starting putget"
  /system/sdcard/bin/busybox nohup /bin/sh /root/getput.sh "$HOSTNAME" 2>/root/getput.log >/dev/null &
  PID=$!
  echo $PID > $PIDFILE
  echo $PID
}

stop()
{
  PID=$(cat $PIDFILE)
  kill -9 $PID
  rm $PIDFILE
}

if [ $# -eq 0 ]; then
  start
else
  case $1 in start|stop|status)
    $1
    ;;
  esac
fi
