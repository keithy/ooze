# Ooze - Server Mobility System

## Original Design Request

Develop a system for server mobility called **ooze**.

**Concept**: A ZFS filesystem can be sent to another server, enabling workload migration between hosts.

**Architecture** (following the sensible pattern):

- `ooze` - Dumb wrapper command (like sensible)
- `ooze-zfs` - Implementation for porting a ZFS volume structure from server A to server B
- `ooze-nixos` - Command for NixOS on root implementation (most likely target OS)

## Key Design Decisions

1. **Wrapper Pattern**: Follow the sensible model - `ooze` is a dumb wrapper that delegates to subcommands
2. **ZFS Send/Receive**: Core mechanism for filesystem migration via `zfs send -R | ssh | zfs receive`
3. **NixOS Target**: Primary target OS is NixOS on root (ZFS implementation)
4. **Dataset Hierarchy**: Support for structured ZFS layouts with OS, safe, and data datasets

## References

- [sensible](https://github.com/keithy/sensible) - Remote execution pattern
- [angelbox/nixbox](https://github.com/keithy/angelbox) - Existing NixOS+ZFS provisioning
