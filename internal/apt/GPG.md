# GPG

Verify Debian Bullseye Stable with public keys: http://ftp.debian.org/debian/dists/bullseye/InRelease

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

gpg --import *.asc

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


Verify Grafana InRelease without adding the public key:

```
root@bullseye:/tmp# gpg --verify /vagrant/GrafanaInRelease
gpg: Signature made Wed Apr 28 11:33:09 2021 UTC
gpg:                using RSA key 4E40DDF6D76E284A4A6780E48C8C34C524098CB6
gpg: Can't check signature: No public key
root@bullseye:/tmp# echo $?
2
```


Verify Grafana InRelease with the public key:

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


## Extract file

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
