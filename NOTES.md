# Notes

This document contains various commands run during development.

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


## Output

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
