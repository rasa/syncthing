#!/usr/bin/env make

# http://security.ubuntu.com/ubuntu/pool/universe/p/python2.7/?C=M;O=D

ROOT := http://security.ubuntu.com/ubuntu/pool/universe/p/python2.7

DEBS := python2.7-minimal_2.7.18-13ubuntu1.5_amd64.deb \
	libpython2.7-minimal_2.7.18-13ubuntu1.5_amd64.deb \
	libpython2.7-stdlib_2.7.18-13ubuntu1.5_amd64.deb

help:
	@echo "Type: 'make install' to install python2.7 on Ubuntu 23.10 or greater"

install: $(DEBS)
	sudo dpkg -i $(DEBS)

python2.7-minimal_2.7.18-13ubuntu1.5_amd64.deb:
	wget $(ROOT)/$@

libpython2.7-minimal_2.7.18-13ubuntu1.5_amd64.deb:
	wget $(ROOT)/$@

libpython2.7-stdlib_2.7.18-13ubuntu1.5_amd64.deb:
	wget $(ROOT)/$@

clean:
	rm -f $(DEBS)

.PHONY: help install clean
