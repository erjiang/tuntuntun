tuntuntun
=========

The goal of **tuntuntun** (t^3 or t cubed or tubed) is to implement an
abstraction on top of ICMP/TCP/UDP in order to multiplex multiple network
connections into one connection at the packet level. That is, with a local
machine will be able to utilize multiple connections to surf the Internet with
the assistance of a remote machine also using tuntuntun.

Use case
--------

If you have a mobile computer with two unreliable wireless connections (e.g.
tethered to a cell phone and connected to a public Wi-Fi AP), you can combine
those two connections into one more reliable connection by proxying through a
remote server. Packets will automatically flow through one or both of the
connections, and automatically switch connections if it suffers high latency
on one connection. Or, if both connections are fairly stable, then you can
have a total bandwidth that is the sum of the two connections'.

Milestones
----------

1. *(done)* Convince Go to open a tun device and read from it.
2. Be able to proxy packets on the local machine by rewriting sender IP.
3. Auto-configure routing to route all network traffic through the tun device.
4. Handle retransmission, ordering, etc. (Effectively reimplementing TCP.)
5. Load-balance connection across multiple physical connections.
