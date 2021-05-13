# Notes

This document contains various commands run during development.

## DPKG

This section contains various commands used to detect local changes when installing a package using `dpkg`.

### Installing

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


### Unpacking

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


## APT

This section contains various APT command outputs.

### Nomimal

Default update:

```
root@buster:/home/vagrant# apt update
Hit:1 http://deb.debian.org/debian buster InRelease
Hit:2 http://security.debian.org/debian-security buster/updates InRelease
Hit:3 http://deb.debian.org/debian buster-updates InRelease
Hit:4 http://deb.debian.org/debian buster-backports InRelease
Reading package lists... Done
Building dependency tree
Reading state information... Done
13 packages can be upgraded. Run 'apt list --upgradable' to see them.
```

Adding a new source:

```
root@buster:/home/vagrant# wget -q -O - https://packages.grafana.com/gpg.key | sudo apt-key add -
OK
root@buster:/home/vagrant# echo "deb https://packages.grafana.com/oss/deb stable main" | sudo tee -a /etc/apt/sources.list.d/grafana.list
```

Update with a new source:

```
root@buster:/home/vagrant# apt update
Hit:1 http://deb.debian.org/debian buster InRelease
Hit:2 http://deb.debian.org/debian buster-updates InRelease
Hit:3 http://security.debian.org/debian-security buster/updates InRelease
Hit:4 http://deb.debian.org/debian buster-backports InRelease
Get:5 https://packages.grafana.com/oss/deb stable InRelease [12.1 kB]
Get:6 https://packages.grafana.com/oss/deb stable/main amd64 Packages [21.9 kB]
  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY 8C8C34C524098CB6
Reading package lists... Done
W: GPG error: https://packages.grafana.com/oss/deb stable InRelease: The following signatures couldn't be verified because the public key is not available: NO_PUBKEY 8C8C34C524098CB6
E: The repository 'https://packages.grafana.com/oss/deb stable InRelease' is not signed.
N: Updating from such a repository can't be done securely, and is therefore disabled by default.
N: See apt-secure(8) manpage for repository creation and user configuration details.
```

Installing a new package:

```
root@buster:/home/vagrant# apt-get install grafana
Reading package lists... Done
Building dependency tree
Reading state information... Done
The following additional packages will be installed:
  fontconfig-config fonts-dejavu-core libfontconfig1
The following NEW packages will be installed:
  fontconfig-config fonts-dejavu-core grafana libfontconfig1
0 upgraded, 4 newly installed, 0 to remove and 13 not upgraded.
Need to get 54.0 MB of archives.
After this operation, 182 MB of additional disk space will be used.
Do you want to continue? [Y/n] y
Get:1 https://packages.grafana.com/oss/deb stable/main amd64 grafana amd64 7.5.5 [52.3 MB]
Get:2 http://deb.debian.org/debian buster/main amd64 fonts-dejavu-core all 2.37-1 [1068 kB]
Get:3 http://deb.debian.org/debian buster/main amd64 fontconfig-config all 2.13.1-2 [280 kB]
Get:4 http://deb.debian.org/debian buster/main amd64 libfontconfig1 amd64 2.13.1-2 [346 kB]
Fetched 54.0 MB in 1s (42.2 MB/s)
perl: warning: Setting locale failed.
perl: warning: Please check that your locale settings:
	LANGUAGE = (unset),
	LC_ALL = (unset),
	LC_CTYPE = "UTF-8",
	LC_TERMINAL = "iTerm2",
	LANG = "C.UTF-8"
    are supported and installed on your system.
perl: warning: Falling back to a fallback locale ("C.UTF-8").
locale: Cannot set LC_CTYPE to default locale: No such file or directory
locale: Cannot set LC_ALL to default locale: No such file or directory
Preconfiguring packages ...
Selecting previously unselected package fonts-dejavu-core.
(Reading database ... 28958 files and directories currently installed.)
Preparing to unpack .../fonts-dejavu-core_2.37-1_all.deb ...
Unpacking fonts-dejavu-core (2.37-1) ...
Selecting previously unselected package fontconfig-config.
Preparing to unpack .../fontconfig-config_2.13.1-2_all.deb ...
Unpacking fontconfig-config (2.13.1-2) ...
Selecting previously unselected package libfontconfig1:amd64.
Preparing to unpack .../libfontconfig1_2.13.1-2_amd64.deb ...
Unpacking libfontconfig1:amd64 (2.13.1-2) ...
Selecting previously unselected package grafana.
Preparing to unpack .../grafana_7.5.5_amd64.deb ...
Unpacking grafana (7.5.5) ...
Setting up fonts-dejavu-core (2.37-1) ...
Setting up fontconfig-config (2.13.1-2) ...
locale: Cannot set LC_CTYPE to default locale: No such file or directory
locale: Cannot set LC_ALL to default locale: No such file or directory
Setting up libfontconfig1:amd64 (2.13.1-2) ...
Setting up grafana (7.5.5) ...
Adding system user `grafana' (UID 108) ...
Adding new user `grafana' (UID 108) with group `grafana' ...
Not creating home directory `/usr/share/grafana'.
### NOT starting on installation, please execute the following statements to configure grafana to start automatically using systemd
 sudo /bin/systemctl daemon-reload
 sudo /bin/systemctl enable grafana-server
### You can start grafana-server by executing
 sudo /bin/systemctl start grafana-server
Processing triggers for systemd (241-7~deb10u7) ...
Processing triggers for man-db (2.8.5-2) ...
Processing triggers for libc-bin (2.28-10) ...
```

### Missing package

```
root@buster:/home/vagrant# apt-get install grafana-oss
Reading package lists... Done
Building dependency tree
Reading state information... Done
E: Unable to locate package grafana-oss
```

### Invalid GPG key

```
root@buster:/home/vagrant# apt update
Hit:1 http://deb.debian.org/debian buster InRelease
Hit:2 http://deb.debian.org/debian buster-updates InRelease
Hit:3 http://security.debian.org/debian-security buster/updates InRelease
Hit:4 http://deb.debian.org/debian buster-backports InRelease
Get:5 https://packages.grafana.com/oss/deb stable InRelease [12.1 kB]
Err:5 https://packages.grafana.com/oss/deb stable InRelease
  The following signatures couldn't be verified because the public key is not available: NO_PUBKEY 8C8C34C524098CB6
Reading package lists... Done
W: GPG error: https://packages.grafana.com/oss/deb stable InRelease: The following signatures couldn't be verified because the public key is not available: NO_PUBKEY 8C8C34C524098CB6
E: The repository 'https://packages.grafana.com/oss/deb stable InRelease' is not signed.
N: Updating from such a repository can't be done securely, and is therefore disabled by default.
N: See apt-secure(8) manpage for repository creation and user configuration details.
```

### Invalid URI

```
root@buster:/home/vagrant# apt update
Hit:1 https://packages.grafana.com/oss/deb stable InRelease
Err:2 https://packagesssss.grafana.com/oss/deb stable InRelease
  Could not resolve 'packagesssss.grafana.com'
Hit:3 http://security.debian.org/debian-security buster/updates InRelease
Hit:4 http://deb.debian.org/debian buster InRelease
Hit:5 http://deb.debian.org/debian buster-updates InRelease
Hit:6 http://deb.debian.org/debian buster-backports InRelease
```

### Invalid DPKG status file

```
Reading package lists... Error!
E: Encountered a section with no Package: header
E: Problem with MergeList /var/lib/dpkg/status
E: The package lists or status file could not be parsed or opened.
```


## GPG

### Verify

How to verify Debian Bullseye Stable with public keys: http://ftp.debian.org/debian/dists/bullseye/InRelease

1. Download the keys

```
wget https://ftp-master.debian.org/keys/archive-key-9-security.asc
wget https://ftp-master.debian.org/keys/archive-key-10-security.asc
wget https://ftp-master.debian.org/keys/archive-key-11-security.asc

wget https://ftp-master.debian.org/keys/release-9.asc
wget https://ftp-master.debian.org/keys/release-10.asc
wget https://ftp-master.debian.org/keys/release-11.asc

wget https://ftp-master.debian.org/keys/archive-key-9.asc
wget https://ftp-master.debian.org/keys/archive-key-10.asc
wget https://ftp-master.debian.org/keys/archive-key-11.asc
```

2. Import the keys

```
gpg --import *.asc
```

3. Verify the InRelease file

```
root@bullseye:~# gpg --verify InRelease
gpg: Signature made Tue May  4 14:10:42 2021 UTC
gpg:                using RSA key 16E90B3FDF65EDE3AA7F323C04EE7237B7D453EC
gpg: Good signature from "Debian Archive Automatic Signing Key (9/stretch) <ftpmaster@debian.org>" [unknown]
gpg: WARNING: This key is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: E1CF 20DD FFE4 B89E 8026  58F1 E0B1 1894 F66A EC98
     Subkey fingerprint: 16E9 0B3F DF65 EDE3 AA7F  323C 04EE 7237 B7D4 53EC
gpg: Signature made Tue May  4 14:10:43 2021 UTC
gpg:                using RSA key 0146DC6D4A0B2914BDED34DB648ACFD622F3D138
gpg: Good signature from "Debian Archive Automatic Signing Key (10/buster) <ftpmaster@debian.org>" [unknown]
gpg: WARNING: This key is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: 80D1 5823 B7FD 1561 F9F7  BCDD DC30 D7C2 3CBB ABEE
     Subkey fingerprint: 0146 DC6D 4A0B 2914 BDED  34DB 648A CFD6 22F3 D138
```

Repeat the same test with Grafana:

* Verify Grafana InRelease without adding the public key:

```
root@bullseye:/tmp# gpg --verify /vagrant/GrafanaInRelease
gpg: Signature made Wed Apr 28 11:33:09 2021 UTC
gpg:                using RSA key 4E40DDF6D76E284A4A6780E48C8C34C524098CB6
gpg: Can't check signature: No public key
root@bullseye:/tmp# echo $?
2
```

* Verify Grafana InRelease with the public key:

```
wget https://packages.grafana.com/gpg.key
mv gpg.key grafana.key
gpg --import grafana.key
gpg: key 8C8C34C524098CB6: public key "Grafana <info@grafana.com>" imported
gpg: Total number processed: 1
gpg:               imported: 1

root@bullseye:~# gpg --verify /vagrant/GrafanaInRelease
gpg: Signature made Wed Apr 28 11:33:09 2021 UTC
gpg:                using RSA key 4E40DDF6D76E284A4A6780E48C8C34C524098CB6
gpg: Good signature from "Grafana <info@grafana.com>" [unknown]
gpg: WARNING: This key is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: 4E40 DDF6 D76E 284A 4A67  80E4 8C8C 34C5 2409 8CB6
root@bullseye:~# echo $?
0
```

### Extract

How to extract the content of a GPG clearsigned document.

```
$  head InRelease
-----BEGIN PGP SIGNED MESSAGE-----
Hash: SHA256

Origin: Debian
Label: Debian
Suite: testing
Codename: bullseye
Changelogs: https://metadata.ftp-master.debian.org/changelogs/@CHANGEPATH@_changelog
Date: Tue, 04 May 2021 14:09:27 UTC
Valid-Until: Tue, 11 May 2021 14:09:27 UTC

$ gpg --output Release InRelease

$ head Release
Origin: Debian
Label: Debian
Suite: testing
Codename: bullseye
Changelogs: https://metadata.ftp-master.debian.org/changelogs/@CHANGEPATH@_changelog
Date: Tue, 04 May 2021 14:09:27 UTC
Valid-Until: Tue, 11 May 2021 14:09:27 UTC
Acquire-By-Hash: yes
No-Support-for-Architecture-all: Packages
Architectures: all amd64 arm64 armel armhf i386 mips64el mipsel ppc64el s390x
```
