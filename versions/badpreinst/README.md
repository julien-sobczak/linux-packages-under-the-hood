# Wrong preinst script

```
# cat > debian/DEBIAN/preinst <<END
#!/bin/bash
exit 1;
END
# chmod 0755 debian/DEBIAN/preinst
# dpkg-deb --build debian hello_1.1-1_amd64.deb
dpkg-deb: building package 'hello' in 'hello_1.1-1_amd64.deb'.
```

When installing the package:

```
# dpkg -i hello_1.1-1_amd64.deb
(Reading database ... 27404 files and directories currently installed.)
Preparing to unpack hello_1.1-1_amd64.deb ...
dpkg: error processing archive hello_1.1-1_amd64.deb (--install):
 new hello package pre-installation script subprocess returned error exit status 1
Errors were encountered while processing:
 hello_1.1-1_amd64.deb
```

Path: `archives.c#archivefiles` > `unpack.c#process_archive` > `script.c#maintscript_new` > `script.c#maintscript_exec` > `subproc.c#subproc_reap` > `subproc.c#subproc_check`
