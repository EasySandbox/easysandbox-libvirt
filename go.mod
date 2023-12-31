module git.voidnet.tech/kev/easysandbox-livbirt

go 1.19

require (
	git.voidnet.tech/kev/easysandbox v0.0.0
	github.com/adrg/xdg v0.4.0
	github.com/estebangarcia21/subprocess v0.0.0-20230526204252-a1a6de4773be
	libvirt.org/go/libvirt v1.9004.0
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace git.voidnet.tech/kev/easysandbox => ../easysandbox

replace github.com/EgosOwn/guestfs => ../libguestfs-1.50.0/golang/src/libguestfs.org/guestfs
