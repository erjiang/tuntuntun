#!/bin/bash

ORIG_GATEWAY=`route -n | grep '^0\.0\.0\.0' | awk '{print $2}'`
PPP_REMOTE=192.168.7.2

control_c()
# run if user hits control-c
{
    echo Restoring gateway to $ORIG_GATEWAY...
    route del default
    route add default gw $ORIG_GATEWAY
    exit $?
}
 
# trap keyboard interrupt (control-c)
trap control_c SIGINT

echo Setting gateway from $ORIG_GATEWAY to $PPP_REMOTE
route del default
route add default gw $PPP_REMOTE

echo Ctrl-C to exit and restore gateway.

while true; do
    sleep 10;
done;
