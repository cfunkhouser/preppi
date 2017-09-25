# PrepPi

PrepPi is a simple binary designed to run early in the boot process of a
Raspberry Pi (and specifically, the Raspbian OS) which copies files from one
location to another.

The intended use is to copy files from a directory under `/boot` - which in
Raspbian is the mountpoint of the FAT-formatted boot partition used by the
Rapsberry Pi bootloader - to arbitrary locations inside of the OS partition. The
rationale is that almost any OS which can write an SD card (and certainly all of
the most prevalent desktop OSes) can write to a FAT volume, while relatively
few can write to EXT4 volumes, so providing a simple mechanism to configure the
OS via writes to the FAT volume is highly desirable.

The secondary motivation for this approach is the idea that editing config files
is easy, and since updating a binary in an OS image is relatively difficult, we
shouldn't encode the logic for configuration in a binary. For this reason, the
`preppi` binary itself is fairly simple, and should remain that way. Any
complicated logic for configuration (ie, template expansion, macros, etc) should
exist elsewhere, and `preppi` should only concern itself with the output of such
a tool.

The `.deb` package produced by the build scripts includes a `systemd` service
which executes `preppi` after local filesystems are mounted but before the
system enters multi-user mode to minimize the chances of the configs having 
already been read. This means that very-early boot configuration is still
impossible with PrepPi.

## Using PrepPi

1.  Flash an OS image with PrepPi installed to the SD card
1.  Mount the boot volume locally
1.  Create a directory `preppi/` and file `preppi/preppi.conf`, and populate
    `preppi.conf`
1.  Boot the OS - PrepPi will place the files where you specify

### Example `preppi.conf`

To set the hostname of the system, the PrepPi configuration might be:

```json
{
  "map": [
    {
      "source": "/boot/preppi/etc-hostname",
      "destination": "/etc/hostname",
      "mode": 420,
      "dirmode": 493,
      "uid": 0,
      "gid": 0,
      "clobber": true
    },
    {
      "source": "/boot/preppi/etc-hosts",
      "destination": "/etc/hosts",
      "mode": 420,
      "dirmode": 493,
      "uid": 0,
      "gid": 0,
      "clobber": true
    }
  ]
}
```

Note that the `mode` and `dirmode` are standard unix file modes, expressed as
decimal values. This is because JSON parsers - and specifically, the Go JSON
implementation - don't handle octal very well. In this case, `420 = 0644` and
`493 = 0755`.

## Versions

The versions and notable changes are listed below.

### `v0.1.1` - 2017-09-25
-   Refactored for subcommands
-   Now includes a `bake` mode, which generates configs from templates

### `v0.1` - 2017-08-28
-   Basic functionality
    -   Applies files to FS, and then reboots the system to force changes to
        take effect
    -   UID/GID enforcement works, but is not unit tested or checked due to
        FS abstraction shortcomings

## Planned Future Features
-   Early boot configuration (ie, filesystem mounting)
-   Package installation / Script execution

## Contributors
-   Christian Funkhouser ([cfunkhouser](http://github.com/cfunkhouser))

## Alternatives
-   [device-init](http://github.com/hypriot/device-init) - The `device-init`
    package is well established and aims to accomplish a similar goal, but with
    configuration expansion in the binary.