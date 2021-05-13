#!/bin/bash
#
# This script must be run inside the Vagrant virtual machine:
#
#   vagrant up
#   vagrant ssh
#   vagrant:~# /vagrant/init.sh

echo "Initializing the virtual machine..."

sudo apt update

# Install various packages
sudo apt install vim binutils fakeroot gnupg

# Build Debian demo packages
cd /vagrant/hello
dpkg-deb --build -Z none 1.1-1 hello_1.1-1_amd64.deb
dpkg-deb --build -Z none 2.1-1 hello_2.1-1_amd64.deb
dpkg-deb --build -Z none 3.1-1 hello_3.1-1_amd64.deb
# NOTE Use dpkg-deb instead of dpkg to use the -Z option to disable compression to make Golang code simpler.
cd -

echo "You are all done!"
