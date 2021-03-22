# Correct package with maintainer scripts

```
# dpkg --build debian hello_1.1-1_amd64.deb
# dpkg -i hello_1.1-1_amd64.deb
Selecting previously unselected package hello.
(Reading database ... 27403 files and directories currently installed.)
Preparing to unpack hello_1.1-1_amd64.deb ...
preinst hello
Unpacking hello (1.1-1) ...
Setting up hello (1.1-1) ...
postinst hello
```

The output of maintainer scripts is displayed in the console.
