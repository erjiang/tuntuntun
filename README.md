tuntuntun
=========

The goal of **tuntuntun** (or "ttt") is to implement an
abstraction on top of IPv4 in order to multiplex multiple network
connections into one connection at the packet level. That is, a local
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
2. *(done)* Be able to proxy packets on the local machine by rewriting sender IP.
3. Add basic security features.
4. Auto-configure routing to route all network traffic through the tun device.
5. Handle retransmission, ordering, etc. (Effectively reimplementing TCP.)
6. *(done)* Load-balance connection across multiple physical connections.
7. Basic automatic failover strategy.
8. Configurable load-balancing and failover strategies.


Questions and Answers
---------------------

Q: Does this offer me any security?

A: No, it offers you no real security. If you want security, you can use a
"real" VPN (like OpenVPN) in conjunction with tuntuntun. In fact, the current
implementation is so insecure that you shouldn't run it on an
Internet-connected machine.

Q: Why not just use OpenVPN + ifenslave to bond two interfaces?

A: Putting more of the load-balancing logic into this software makes it easier
to experiment with different load-balancing and/or failover strategies. For
example, you can modify the code to send 75% of packets over one link, and the
rest on another link, or do instant failover when one link's latency exceeds a
certain threshold.

Q: How does this compare to programs that don't require a proxy server?

A: Many connection-sharing/load-balancing programs operate by assigning
different connections to different links. For example, if you are downloading
two files, each file can be downloaded over a different link. The advantage is
that no outside proxy server is needed. But, in this strategy, it's not possible
to speed up one connection using multiple links.
