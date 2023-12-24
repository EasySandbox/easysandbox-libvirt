# libvirt sandboxes and xpra

Libvirt is a bit tricky with xpra, configuring a user a NAT or bridged network
is quite complex and modfies the system heavily, and requiring the user to configure it
is asking a lot.

Another solution is to use a reverse tcp connection.
A broker service is ran on the host that listens at a predefined point, and when contacted
by a VM it executes a new shootback service which listens on two ports, one for the VM and
another for the host's xpra client.

1. When a libvirt vm is started, a shootback service is started alongside it on a particular port
2. the port is set in the VM's home image in /tmp/easysandboxguiport
3. a shootback client is started against that port, forwarding traffic to the VM's xpra server
