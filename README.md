# Linux Packages Under the Hood

## Installing

```
$ vagrant up
$ vagrant ssh
$ sudo apt update
$ sudo apt install vim binutils fakeroot
```


## dpkg-list

```
# From the host
$ make build-dpkg
# From the VM
# /vagrant/bin/dpkg --install /vagrant/hello.deb
```

## Building the package `hello`

This demo package is present under the directory `hello`, which contains various versions of this package.

## Building

```sh
$ cd ./hello/
$ dpkg-deb --build -Z none 1.1-1 hello_1.1-1_amd64.deb
$ dpkg-deb --build -Z none 2.1-1 hello_2.1-1_amd64.deb
# NOTE Use dpkg-deb instead of dpkg to use the -Z option to disable compression to make Golang code simpler.
```

## Installing

```sh
# Install version 1.1-1
root@bullseye:/home/vagrant# mkdir /root/diff
root@bullseye:/home/vagrant# cp -R /var/lib/dpkg /root/diff/v1
root@bullseye:/home/vagrant# cd /vagrant/hello
root@bullseye:/vagrant/hello# dpkg -i hello_1.1-1_amd64.deb
Selecting previously unselected package hello.
(Reading database ... 27403 files and directories currently installed.)
Preparing to unpack hello_1.1-1_amd64.deb ...
preinst says hello
Unpacking hello (1.1-1) ...
Setting up hello (1.1-1) ...
postinst says hello
root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/v2
root@bullseye:/vagrant/hello# diff -r /root/diff/v{1,2}
Only in /root/diff/v2/info: hello.list
Only in /root/diff/v2/info: hello.md5sums
Only in /root/diff/v2/info: hello.postinst
Only in /root/diff/v2/info: hello.postrm
Only in /root/diff/v2/info: hello.preinst
Only in /root/diff/v2/info: hello.prerm
diff -r /root/diff/v1/status /root/diff/v2/status
1508a1509,1517
> Package: hello
> Status: install ok installed
> Priority: optional
> Section: base
> Maintainer: Julien Sobczak
> Architecture: amd64
> Version: 1.1-1
> Description: Say Hello
>
diff -r /root/diff/v1/status-old /root/diff/v2/status-old
436c436
< Status: install ok unpacked
---
> Status: install ok installed
454c454
< Status: install ok unpacked
---
> Status: install ok installed
472c472
< Status: install ok unpacked
---
> Status: install ok installed
1120c1120
< Status: install ok unpacked
---
> Status: install ok installed
2026c2026
< Status: install ok unpacked
---
> Status: install ok installed
2146c2146
< Status: install ok triggers-pending
---
> Status: install ok installed
2173d2172
< Triggers-Pending: ldconfig
2367c2366
< Status: install ok unpacked
---
> Status: install ok installed
2385c2384
< Status: install ok unpacked
---
> Status: install ok installed
2667c2666
< Status: install ok unpacked
---
> Status: install ok installed
2680c2679
<  /etc/ld.so.conf.d/fakeroot-x86_64-linux-gnu.conf newconffile
---
>  /etc/ld.so.conf.d/fakeroot-x86_64-linux-gnu.conf f9f2331782e9078d5472c77e1d9cd869
3033c3032
< Status: install ok unpacked
---
> Status: install ok installed
5263c5262
< Status: install ok triggers-pending
---
> Status: install ok installed
5293d5291
< Triggers-Pending: /usr/share/applications /usr/lib/mime/packages
5296c5294
< Status: install ok triggers-pending
---
> Status: install ok installed
5323d5320
< Triggers-Pending: /usr/share/man
6997c6994
< Status: install ok unpacked
---
> Status: install ok installed
7021c7018
< Status: install ok unpacked
---
> Status: install ok installed
7030d7026
< Config-Version: 2:8.2.2434-1
7044c7040
< Status: install ok unpacked
---
> Status: install ok installed
7067c7063
< Status: install ok unpacked
---
> Status: install ok installed
7075d7070
< Config-Version: 2:8.2.2434-1

# Install version 1.1-2
root@bullseye:/vagrant/hello# dpkg -i hello_1.1-2_amd64.deb
(Reading database ... 27404 files and directories currently installed.)
Preparing to unpack hello_1.1-2_amd64.deb ...
prerm says hello
preinst says hello
Unpacking hello (1.1-2) over (1.1-1) ...
postrm says hello
Setting up hello (1.1-2) ...
postinst says hello
root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/v3
root@bullseye:/vagrant/hello# diff -r /root/diff/v{2,3}
diff -r /root/diff/v2/info/hello.md5sums /root/diff/v3/info/hello.md5sums
1c1
< 774d9d8215923c327f9b324ffee0d36b  usr/bin/hello
---
> 1baff99e1fb599cf1837e2761fe239b4  usr/bin/hello
diff -r /root/diff/v2/status /root/diff/v3/status
1515c1515
< Version: 1.1-1
---
> Version: 1.1-2
diff -r /root/diff/v2/status-old /root/diff/v3/status-old
1508a1509,1517
> Package: hello
> Status: install ok installed
> Priority: optional
> Section: base
> Maintainer: Julien Sobczak
> Architecture: amd64
> Version: 1.1-1
> Description: Say Hello
>

# Install version 1.2-1
root@bullseye:/vagrant/hello# dpkg -i hello_1.2-1_amd64.deb
(Reading database ... 27404 files and directories currently installed.)
Preparing to unpack hello_1.2-1_amd64.deb ...
prerm says hello
preinst says hello
Unpacking hello (1.2-1) over (1.1-2) ...
postrm says hello
Setting up hello (1.2-1) ...

Configuration file '/usr/bin/hello'
 ==> File on system created by you or by a script.
 ==> File also in package provided by package maintainer.
   What would you like to do about it ?  Your options are:
    Y or I  : install the package maintainer's version
    N or O  : keep your currently-installed version
      D     : show the differences between the versions
      Z     : start a shell to examine the situation
 The default action is to keep your current version.
*** hello (Y/I/N/O/D/Z) [default=N] ? I
Installing new version of config file /usr/bin/hello ...
postinst says hello
root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/v4
root@bullseye:/vagrant/hello# diff -r /root/diff/v{3,4}
Only in /root/diff/v4/info: hello.conffiles
diff -r /root/diff/v3/info/hello.md5sums /root/diff/v4/info/hello.md5sums
1c1
< 1baff99e1fb599cf1837e2761fe239b4  usr/bin/hello
---
> 8884f49ed9b78e54695a97a6a24518da  usr/bin/hello
diff -r /root/diff/v3/status /root/diff/v4/status
1515c1515,1517
< Version: 1.1-2
---
> Version: 1.2-1
> Conffiles:
>  /usr/bin/hello 8884f49ed9b78e54695a97a6a24518da
diff -r /root/diff/v3/status-old /root/diff/v4/status-old
1515c1515
< Version: 1.1-1
---
> Version: 1.1-2
root@bullseye:/vagrant/hello# cat > /usr/bin/hello <<EOF
#!/bin/env python3

print("Hello world")
EOF

# Install version 1.2-2
root@bullseye:/vagrant/hello# dpkg -i hello_1.2-2_amd64.deb
(Reading database ... 27404 files and directories currently installed.)
Preparing to unpack hello_1.2-2_amd64.deb ...
prerm says hello
preinst says hello
Unpacking hello (1.2-2) over (1.2-1) ...
postrm says hello
Setting up hello (1.2-2) ...

Configuration file '/usr/bin/hello'
 ==> Modified (by you or by a script) since installation.
 ==> Package distributor has shipped an updated version.
   What would you like to do about it ?  Your options are:
    Y or I  : install the package maintainer's version
    N or O  : keep your currently-installed version
      D     : show the differences between the versions
      Z     : start a shell to examine the situation
 The default action is to keep your current version.
*** hello (Y/I/N/O/D/Z) [default=N] ? D
--- /usr/bin/hello      2021-03-27 15:30:25.983395104 +0000
+++ /usr/bin/hello.dpkg-new     2021-03-27 15:16:37.000000000 +0000
@@ -1,3 +1,3 @@
 #!/bin/env python3

-print("Hello world")
+print("hello world!")

Configuration file '/usr/bin/hello'
 ==> Modified (by you or by a script) since installation.
 ==> Package distributor has shipped an updated version.
   What would you like to do about it ?  Your options are:
    Y or I  : install the package maintainer's version
    N or O  : keep your currently-installed version
      D     : show the differences between the versions
      Z     : start a shell to examine the situation
 The default action is to keep your current version.
*** hello (Y/I/N/O/D/Z) [default=N] ? O
postinst says hello
root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/v5
root@bullseye:/vagrant/hello# diff -r /root/diff/v{4,5}
diff -r /root/diff/v4/info/hello.md5sums /root/diff/v5/info/hello.md5sums
1c1
< 8884f49ed9b78e54695a97a6a24518da  usr/bin/hello
---
> 340b62fe077a4424b05dd104865635b3  usr/bin/hello
diff -r /root/diff/v4/status /root/diff/v5/status
1515c1515
< Version: 1.2-1
---
> Version: 1.2-2
1517c1517
<  /usr/bin/hello 8884f49ed9b78e54695a97a6a24518da
---
>  /usr/bin/hello 340b62fe077a4424b05dd104865635b3
diff -r /root/diff/v4/status-old /root/diff/v5/status-old
1515c1515,1517
< Version: 1.1-2
---
> Version: 1.2-1
> Conffiles:
>  /usr/bin/hello 8884f49ed9b78e54695a97a6a24518da
```


## Unpacking

```sh
root@bullseye:/vagrant/hello# dpkg --unpack hello_2.1-1_amd64.deb
Selecting previously unselected package hello.
(Reading database ... 25063 files and directories currently installed.)
Preparing to unpack hello_2.1-1_amd64.deb ...
preinst says hello
Unpacking hello (1.2-2) ...

root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/dpkg2

root@bullseye:/vagrant/hello# diff -r /root/diff/dpkg{1,2}
Only in /root/diff/dpkg2/info: hello.conffiles
Only in /root/diff/dpkg2/info: hello.list
Only in /root/diff/dpkg2/info: hello.md5sums
Only in /root/diff/dpkg2/info: hello.postinst
Only in /root/diff/dpkg2/info: hello.postrm
Only in /root/diff/dpkg2/info: hello.preinst
Only in /root/diff/dpkg2/info: hello.prerm
diff -r /root/diff/dpkg1/status /root/diff/dpkg2/status
1429a1430,1440
> Package: hello
> Status: install ok unpacked
> Priority: optional
> Section: base
> Maintainer: Julien Sobczak
> Architecture: amd64
> Version: 1.2-2
> Conffiles:
>  /etc/hello/settings.conf newconffile
> Description: Say Hello
>

root@bullseye:/vagrant/hello# ls /etc/hello/
settings.conf.dpkg-new
root@bullseye:/vagrant/hello# ls -l /usr/bin/hello
-rwxr-xr-x 1 vagrant vagrant 481 Apr  1 09:58 /usr/bin/hello

################

root@bullseye:/vagrant/hello# dpkg --configure hello
Setting up hello (1.2-2) ...
postinst says hello
root@bullseye:/vagrant/hello# ls /etc/hello/
settings.conf
root@bullseye:/vagrant/hello# ls -l /usr/bin/hello
-rwxr-xr-x 1 vagrant vagrant 481 Apr  1 09:58 /usr/bin/hello

root@bullseye:/vagrant/hello# cp -R /var/lib/dpkg/ /root/diff/dpkg3

root@bullseye:/vagrant/hello# diff -r /root/diff/dpkg{2,3}
diff -r /root/diff/dpkg2/status /root/diff/dpkg3/status
1431c1431
< Status: install ok unpacked
---
> Status: install ok installed
1438c1438
<  /etc/hello/settings.conf newconffile
---
>  /etc/hello/settings.conf 010db41180e7df10564838edd7d8492e
diff -r /root/diff/dpkg2/status-old /root/diff/dpkg3/status-old
1429a1430,1440
> Package: hello
> Status: install ok unpacked
> Priority: optional
> Section: base
> Maintainer: Julien Sobczak
> Architecture: amd64
> Version: 1.2-2
> Conffiles:
>  /etc/hello/settings.conf newconffile
> Description: Say Hello
>
```
