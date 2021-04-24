# Output

## Nomimal

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


## Errors

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
