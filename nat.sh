#!/bin/bash

# nat.sh wan_if lan_if
# turns on NAT, assuming wan_if is your Internet connection, and lan_if is
# your internal interface.
# based on http://www.revsys.com/writings/quicktips/nat.html
# with more help from http://www.netfilter.org/documentation/HOWTO/NAT-HOWTO-1.html

WAN="$1"
LAN="$2"

# change outgoing packets' src addr to our external IP (whatever it is)
iptables -t nat -A POSTROUTING -o "$WAN" -j MASQUERADE

echo 0 > /proc/sys/net/ipv4/conf/"$WAN"/rp_filter


iptables -A FORWARD -s 192.168.8.0/24 -o "$WAN" -j ACCEPT
# iptables -A FORWARD -i "$WAN" -o "$LAN" -m state --state RELATED,ESTABLISHED -j ACCEPT
# iptables -A FORWARD -i "$LAN" -o "$WAN" -j ACCEPT

echo 1 > /proc/sys/net/ipv4/ip_forward
