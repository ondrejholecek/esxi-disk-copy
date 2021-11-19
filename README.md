# esxi-disk-copy
Copy VMDK files to ESXi hosts over SSH and convert them there.

## Background

For long time I was using `1023856-vmware-vdiskmanager-linux.7.0.1` binary from VMware to upload disk images to ESXi datastore. That worked well up to ESXi 6.0. However, after upgrading the ESXi host to 6.7 this program stopped working (because it wants to use SSLv3 which is not supported on ESXi any more) and I wasn't able to find any workaround (enabling this protocol on ESXi manually doesn't work for me).

Using newer binaries didn't work either, because the one from VDDK 6.7 doesn't have any option to specify certificate thumbprint even though it requires it to connect and binary from VDDK 7.0 doesn't seem to support copying files to remote ESXi at all.

Programming custom tool using the library provided by VMware seems to be impossible (?) because it now requires specifying the exact VM where the disk should be uploaded to, which is something I don't want (my disks are supposed to be "templates" that are backing up multiple VMs created later).

## About this tool

The only sane way is to upload the disk file via SSH (SCP) and convert it to "thin" disk there using `vmkfstools` binary on ESXi host itself.

And that is exactly what this tool does. It uses the same command line parameters that the original tool used, but it:
- Connects to ESXi via SSH
- Uploads the disk file (previously "manually" extracted locally from OVF) via SCP to temporary file in the same datastore
- Executes shell command `vmkfstools -i ... -d thin ...` to clone it to final disk file
- Deletes the previous uploaded temporary file

## Usage

The same basic parameters that were used in original binary are supported (however `-t` is ignored).

Following is the command line that I use:

```
$ ./esxi-disk-copy.linux_amd64 -r /tmp/extracted-from-ovf.vmdk -t 6 -h esxihost.example.com \
                               -u root -f /tmp/password '[datastore] templates/template-name-os.vmdk'
```

Note that SSH access must be explicitly enabled on ESXi host.

Parameters:
- `-r` : local vmdk disk extracted from OVF
- `-t` : ignored, accepted for backwards compatibility with original binary
- `-h` : ESXi host name or IP address
- `-u` : user name on ESXi
- `-f` : file containing SSH password for user
- last parameter is "datastore path" (the program expects that such datastore is mounted in `/vmfs/volumes/` on ESXi host

New (option) parameter just for this program:
- `-tmp-file` : path on the same datastore as final file where SCP copies the disk image and deletes it after converting (by default it is random name in root of datastore directory with unix timestamp included to prevent multiple running instancies overwriting the same file)

## Binaries

"Static" binaries for Linux/MacOS/Windows are available in `binaries` directory of this repository.

You can compile your own version by cloning the repository and running `make` to get binary for your architecture (or `make archs` to get all three binaries cross-compiled).




