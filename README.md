# Linux Packages Under the Hood

This repository is the companion of the blog post [_Linux Packages Under the Hood_](https://www.juliensobczak.com/inspect/2021/05/15/linux-packages-under-the-hood.html).

**DO NOT RUN the code in this repository on your machine.** The `dpkg` and `apt` are basic versions used as a learning tool. Always use a local virtual machine instead so that you can trash it easily.

## Installing

```
$ vagrant up
$ vagrant ssh
vagrant$ sudo su
vagrant# /vagrant/init.sh
```

## Building

The repository contains a `Makefile` to build the Go binaries:

```
$ make test      # Run automated tests
$ make build-cmd # Rebuild the binaries `dpkg` and `apt` under `bin/`
```

## Testing

This repository contains a basic `hello` Debian package under the directory `hello/`. This package installs a script displaying the classic "Hello World" message. Several versions are available to test different features of Debian packages:

* `1.1-1`: Golang version (see `./hello.go`) without any external dependencies.
* `2.1-1`: Python version using a conffile to configure the default language.
* `3.1-1`: Bash version declaring a dependency on `cowsay` to display the message in a more funny way.


### dpkg --build

```
# From the VM
vagrant# cd /vagrant/hello
vagrant# /vagrant/bin/dpkg --build 1.1-1 hello_1.1-1_amd64.deb
(Reading database ... 28590 files and directories currently installed.)
Preparing to unpack hello.deb ...
preinst says hello
Unpacking hello (1.1-1) ...
Setting up hello (1.1-1) ...
postinst says hello
vagrant# hello
hello world
```

### dpkg --install

```
# From the VM
vagrant# /vagrant/bin/dpkg --install /vagrant/hello.deb
vagrant# hello
hello world
```

Uninstall the package for the next step:

```
# From the VM
vagrant# dpkg -r hello
(Reading database ... 27664 files and directories currently installed.)
Removing hello (1.1-1) ...
prerm says hello
postrm says hello
```

### apt --install

```
# From the VM

vagrant# cp /etc/apt/sources.list{,.bak}
vagrant# cat > /etc/apt/sources.list << EOF
deb http://deb.debian.org/debian buster main
deb-src http://deb.debian.org/debian buster main
EOF

vagrant# /vagrant/bin/apt --install /vagrant/hello/hello_3.1-1_amd64.deb
Get:1 http://deb.debian.org/debian stable InRelease [121.5 kB]
Get:2 http://deb.debian.org/debian stable/main amd64 Packages [7.9 MB]
The following additional packages will be installed:
	cowsay
Suggested packages:
	filters cowsay-off
=> [cowsay hello]
Get:1 http://deb.debian.org/debian stable/main cowsay all 3.03+dfsg2-6 [20.9 kB]
(Reading database ... 28590 files and directories currently installed.)
Preparing to unpack cowsay_3.03+dfsg2-6_all.deb ...
Unpacking cowsay (3.03+dfsg2-6) ...
Setting up cowsay (3.03+dfsg2-6) ...
Preparing to unpack hello_3.1-1_amd64.deb ...
Unpacking hello (3.1-1) ...
Setting up hello (3.1-1) ...

vagrant# hello
 _____________
< hello world >
 -------------
        \   ^__^
         \  (oo)\_______
            (__)\       )\/\
                ||----w |
                ||     ||
```
