# Missing dependency

Append `Depends: unknown (>= 2.2.1)` into `debian/DEBIAN/control`.

When installing the package:

```
# dpkg -i hello_1.1-1_amd64.deb
(Reading database ... 27404 files and directories currently installed.)
Preparing to unpack hello_1.1-1_amd64.deb ...
Unpacking hello (1.1-1) over (1.1-1) ...
dpkg: dependency problems prevent configuration of hello:
 hello depends on unknown (>= 2.2.1); however:
  Package unknown is not installed.

dpkg: error processing package hello (--install):
 dependency problems - leaving unconfigured
Errors were encountered while processing:
 hello
```

Path: `archives.c#archivefiles` > `packages.c#process_queue` > `configure.c#deferred_configure` > `configure.c#dependencies_ok` > `packages.c#dependencies_ok` > `packages.c#deppossi_ok_found`
