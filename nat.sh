#!/bin/bash

# nat.sh wan_if lan_if
# turns on NAT, assuming wan_if is your Internet connection, and lan_if is
# your internal interface.
# based on http://www.revsys.com/writings/quicktips/nat.html

WAN="$1"
LAN="$2"

iptables -t nat -A POSTROUTING -o "$WAN" -j MASQUERADE
iptables -A FORWARD -i "$WAN" -o "$LAN" -m state --state RELATED,ESTABLISHED -j ACCEPT
iptables -A FORWARD -i "$LAN" -o "$WAN" -j ACCEPT

echo 1 > /proc/sys/net/ipv4/ip_forward
