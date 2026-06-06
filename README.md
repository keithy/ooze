# Ooze

Server mobility through ZFS filesystem migration.

## Concept

Ooze enables moving running server workloads between hosts by leveraging ZFS send/receive. The system treats a ZFS dataset hierarchy as a portable server unit that can be migrated while preserving all properties (compression, snapshots, properties).

## Architecture

```
ooze/ # Main wrapper (dumb dispatcher)
├── cmd/
│   ├── ooze/                 # Main entry point
│   ├── ooze-zfs/             # ZFS volume migration
│   └── ooze-nixos/           # NixOS root implementation
└── pkg/ooze/                 # Shared library
```

## Commands

### ooze (wrapper)

Dumb wrapper that delegates to subcommands. Follows the sensible pattern.

```bash
ooze zfs send tank/data/db1 root@server2
ooze nixos init tank
```

### ooze-zfs

ZFS filesystem mobility:

```bash
ooze-zfs send <dataset> <target>   # Send dataset to target
ooze-zfs receive <dataset>         # Receive dataset
ooze-zfs snapshot <dataset>       # Create migration snapshot
ooze-zfs list [dataset]           # List datasets
ooze-zfs status <dataset>          # Show migration status
```

### ooze-nixos

NixOS root on ZFS implementation:

```bash
ooze-nixos init <pool> # Initialize pool for NixOS root
ooze-nixos send <profile> <target> # Send NixOS config
ooze-nixos receive <pool>          # Configure this host
ooze-nixos list                    # List available profiles
ooze-nixos switch <profile>       # Switch to profile
```

## ZFS Layout

The expected ZFS dataset hierarchy for server workloads:

```
pool/
├── os/                          # OS datasets (canmount=off)
│   ├── nix                      # /nix/store
│   ├── etc                     # /etc
│   └── var                     # /var
├── safe/                       # User data (canmount=on)
│   └── user/
│       ├── home                # /home
│       └── root                # /root
└── data/                       # Application data
    └── <app>                   # /var/lib/<app>
```

## Migration Flow

1. **Snapshot**: Create a point-in-time snapshot of the dataset
2. **Send**: Stream the snapshot via SSH to target (`zfs send -R | ssh | zfs receive`)
3. **Receive**: Import the stream and restore dataset properties
4. **Verify**: Confirm data integrity and mountpoints

## Installation

```bash
# System install
make install

# User install
make install-user
```

## Building

```bash
make build        # Build all binaries
make test         # Run tests
```

## Related Projects

- [sensible](https://github.com/keithy/sensible) - Remote execution for AI agents (pattern followed here)
- [angelbox/nixbox](https://github.com/keithy/angelbox) - NixOS provisioning with ZFS support

## License

MIT
