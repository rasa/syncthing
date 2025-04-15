#!/usr/bin/make

# @TODO deal with:
# gui: Tweak the Restore Versions modal for better usability (#7972)
# https://github.com/syncthing/syncthing/commit/9efac0f0678890cc0a67a8c603ec4bb1d540d89f#diff-cd84434b22af5511a55fac66ee28f9daeda01a012b0d0a3b9ce6feaa53e1dc6d

# https://patch-diff.githubusercontent.com/raw/syncthing/syncthing/pull/8509.diff

PHONY+=default
default: help ;

# What syncthing is using now:
ANGULAR_VER ?= v1.3.20
ANGULAR_TRANSLATE_VER ?= 2.9.0
ANGULAR_TRANSLATE_LOADER_VER ?= 2.11.0
ANGULAR_UTILS_VER ?= 759009c
BOOTSTRAP_VER ?= v3.3.6
FANCYTREE_VER ?= v2.38.0
JQUERY_VER ?= 2.2.2
MOMENT_VER ?= 2.19.4

# latest versions:
NVM_VER ?= v0.40.2
NPM_VER ?= 2.15.11

# shared:
BOWER_VER ?= 1.3.12
GRUNT_VER ?= 0.4.2
GRUNT_CLI_VER ?= 0.1.13
TOUCH_COOKIE_VER ?= 0.12.1

# angular 1.3.20 versions:
ANGULAR_NODE_VER ?= 0.10.48
REQUEST_VER ?= 2.51.0

# angular-translate 2.9.0's versions:
NODE_VER ?= 4.2.6
GRUNT_EXPRESS_VER ?= 0.2.1
GRUNT_KARMA_VER ?= 0.8.3
KARMA_VER ?= 0.12.31

NVM_DIR := $(shell [ -n "$${XDG_CONFIG_HOME}" ] && printf %s "$${XDG_CONFIG_HOME}/nvm" || printf %s "$${HOME}/.nvm")

PYTHON := $(shell which python2.7)

ifeq (, $(PYTHON))
	$(error "Cannot find python2.7")
	$(error "To install on Ubuntu 23.10+, run: make -f python2.mak install")
endif

DIRS = angular.tmp \
	angular-translate.tmp \
	angular-translate-loader.tmp \
	angularUtils.tmp \
	bootstrap.tmp \
	fancytree.tmp \
	jquery.tmp \
	moment.tmp

$(NVM_DIR)/nvm.sh:
	test -f "$@" || \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/$(NVM_VER)/install.sh | bash

# GitHub doesn't support git:// URLs any more
.git_configured:
	git config --global url."https://github.com/".insteadOf git://github.com/
	touch $@

PHONY+=setup
setup: $(NVM_DIR)/nvm.sh .git_configured ## install nvm, and configure git

# $1 dir
# $2 url
# $3 tag
define checkout
	test -d $1 || git clone -q $2 $1
	# git reset --hard
	# git clean -fdx
	cd $1 && \
		git checkout -q $3
endef

# $1 dir
define common
	# Stop bower from saying:
	#  [?] May bower anonymously report usage statistics to improve the tool over time? (Y/n)
	printf '{"interactive": false}' >$@/.bowerrc
	# change git: URLs to https: as GitHub dropped support
	find $@ -type f -name 'bower.json' -o -name package.json -o -name 'npm-shrinkwrap*.json' | xargs \
		sed -i 's|git://|https://|;'
	# pin all package versions
	find $@ -type f -name 'bower.json' -o -name package.json -o -name 'npm-shrinkwrap*.json' | xargs \
		sed -i 's/"~/"/; s/"^/"/;'
endef

angular.tmp: setup
	$(call checkout,$@,https://github.com/angular/angular.js,$(ANGULAR_VER))
	$(call common,$@)
	# not needed, but it makes sense:
	# sed -i 's|"angular": ">=1.2.26 <=1.5"|"angular": "$(ANGULAR_VER)"' $@/package.json

	# @TODO TRY MOVING TO common
	# 2. shared (but installs older node version):
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm install --default $(ANGULAR_NODE_VER)
	#
	# @TODO TRY MOVING TO common
	# 3. shared:
	# fix: npm ERR! cb() never called!
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g npm@$(NPM_VER)
	#
	# 4a. sort shared (installs grunt, grunt-express, grunt-karma, karma, tough-cookie instead)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install \
			request@$(REQUEST_VER) \
			tough-cookie@$(TOUCH_COOKIE_VER)
	#
	# @TODO TRY MOVING TO common2
	# 6. shared:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install
	#
	# @TODO TRY ADDING 7.
	#
	# @TODO TRY INSTALLING WITHOUT -g
	# 8. shared, but has -g:
	# install version 1.3.12
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g bower
	#
	# @TODO TRY INSTALLING WITHOUT -g
	# grunt shared (but here we install grunt later, with -g)
	# 9. not shared, but has -g:
	# without the version number, 0.4.5 would be installed
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g grunt@$(GRUNT_VER)
	#
	# @TODO TRY INSTALLING WITHOUT -g
	# 10. shared, but has -g:
	# without the version number, 1.5.0 would be installed
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g grunt-cli@$(GRUNT_CLI_VER)
	#
	# @TODO TRY WITH ./node_modules/.bin/bower
	# 11. shared, but uses global install:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		bower install
	#
	# @TODO TRY WITH ./node_modules/.bin/grunt
	# 12a. shared (but uses package instead of build:core):
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		grunt package
	#
	cp -pv $@/build/angular.js          angular
	cp -pv $@/build/angular-sanitize.js angular
	-cp -pv $@/LICENSE angular/angular.js.LICENSE
	-cp -pv $@/LICENSE angular/angular-sanitize.js.LICENSE
	touch $@

angular-translate.tmp: setup
	$(call checkout,$@,https://github.com/angular-translate/angular-translate,$(ANGULAR_TRANSLATE_VER))
	$(call common,$@)
	#
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs -r \
		sed -Ei 's/(.*"grunt":\s*")[^"]+(".*)/\1'"$(GRUNT_VER)"'\2/'
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs \
		sed -Ei 's/(.*"grunt-cli":\s*")[^"]+(".*)/\1'"$(GRUNT_CLI_VER)"'\2/'
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs \
		sed -Ei 's/(.*"grunt-express":\s*")[^"]+(".*)/\1'"$(GRUNT_EXPRESS_VER)"'\2/'
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs \
		sed -Ei 's/(.*"grunt-karma":\s*")[^"]+(".*)/\1'"$(GRUNT_KARMA_VER)"'\2/'
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs \
		sed -Ei 's/(.*"karma":\s*")[^"]+(".*)/\1'"$(KARMA_VER)"'\2/'
	# @TODO TRY REMOVING:
	find $@ -type f -name package.json | xargs \
		sed -Ei 's/(.*"tough-cookie":\s*")[^"]+(".*)/\1'"$(TOUCH_COOKIE_VER)"'\2/'
	#
	# 2. shared (but installs newer node version):
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm install --default $(NODE_VER)
	#
	# 3. shared:
	# fix: npm ERR! cb() never called!
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g npm@$(NPM_VER)
	#
	# @TODO TRY INSTALLING grunt IN STEP 9. AS ABOVE
	# 5b. sorta shared (but other installs request, tough-cookie instead)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install \
			grunt@$(GRUNT_VER) \
			grunt-express@$(GRUNT_EXPRESS_VER) \
			grunt-karma@$(GRUNT_KARMA_VER) \
			karma@$(KARMA_VER) \
			tough-cookie@$(TOUCH_COOKIE_VER)
	#
	# 6. shared:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install
	#
	# 7. not shared, but maybe could be shared:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm shrinkwrap --dev
	#
	# @TODO TRY INSTALLED WITH -g
	# 8. shared, but no -g:
	# install version 1.3.12
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install bower@$(BOWER_VER)
	#
	# @TODO TRY INSTALLED WITH -g
	# 10. shared:
	# without the version number, 1.5.0 would be installed
	# moving this up breaks the build
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt-cli@$(GRUNT_CLI_VER)
	#
	# 11. shared, but uses local install:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/bower install
	#
	# 12b. shared, but uses local install,uses build:core instead of package:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/grunt build:core
	#
	cp -pv $@/dist/angular-translate.js angular
	-cp -pv $@/LICENSE angular/angular-translate.js.LICENSE
	touch $@

angular-translate-loader.tmp:
	$(call checkout,$@,https://github.com/angular-translate/bower-angular-translate-loader-static-files,$(ANGULAR_TRANSLATE_LOADER_VER))
	cp -pv $@/angular-translate-loader-static-files.js angular
	# see https://github.com/angular-translate/bower-angular-translate-loader-static-files?tab=readme-ov-file#license
	-cp -pv angular/angular-translate.js.LICENSE angular/angular-translate-loader-static-files.js.LICENSE
	touch $@

angularUtils.tmp:
	$(call checkout,$@,https://github.com/michaelbromley/angularUtils,$(ANGULAR_UTILS_VER))
	cp -pv $@/src/directives/pagination/dirPagination.js angular/angular-dirPagination.js
	-cp -pv $@/license.txt angular/angular-dirPagination.js.LICENSE
	touch $@

bootstrap.tmp:
	$(call checkout,$@,https://github.com/twbs/bootstrap,$(BOOTSTRAP_VER))
	cp -pv $@/dist/css/bootstrap.css       bootstrap/css
	cp -pv $@/dist/css/bootstrap-theme.css bootstrap/css
	cp -pv $@/dist/js/bootstrap.js         bootstrap/js
	-cp -pv $@/LICENSE                      bootstrap/LICENSE
	touch $@

fancytree.tmp:
	$(call checkout,$@,https://github.com/mar10/fancytree,$(FANCYTREE_VER))
	cp -pv $@/dist/jquery.fancytree-all-deps.js fancytree
	-cp -pv $@/dist/LICENSE.txt fancytree
	touch $@

jquery.tmp:
	$(call checkout,$@,https://github.com/jquery/jquery,$(JQUERY_VER))
	cp -pv $@/dist/jquery.js jquery/jquery-$(JQUERY_VER).js
	-cp -pv $@/LICENSE.txt    jquery/LICENSE
	touch $@

moment.tmp:
	$(call checkout,$@,https://github.com/moment/moment,$(MOMENT_VER))
	cp -pv $@/moment.js moment
	-cp -pv $@/LICENSE   moment
	touch $@

PHONY+=all
all: checkout_lfs download checkout_asis ## Run `make checkout_lfs download checkout_asis`

PHONY+=checkout_lfs
checkout_lfs: ## Set git to checkout all text files with LF line endings (Windows only)
ifeq ($(OS),Windows_NT)
	sed -i -e 's|^vendor.*|vendor/** eol=lf|' ../../../.gitattributes
endif

PHONY+=checkout_asis
checkout_asis: ## Set git to checkout all text files "as is" (Windows only)
ifeq ($(OS),Windows_NT)
	git restore ../../../.gitattributes
endif

PHONY+=download
download: $(DIRS) ## Download all the files

PHONY+=clean
clean: ## Clean only the .tmp directories
	find -maxdepth 1 -type d -name "*.tmp" | xargs rm -fr
	rm -fr *.deb

fmt=\033[36m%-24s\033[0m %s\n
PHONY+=help
help: ## Display a help screen
	@printf '$(fmt)' 'Target' 'Description'
	@printf '$(fmt)' '------' '------------------------------------------'
	@awk -F ':.*##[ \t]*' '/^[^#: \t]+:.*##/ {printf "$(fmt)", $$1, $$2}' $(MAKEFILE_LIST) | sort

PHONY+=nvm
nvm: $(NVM_DIR)/nvm.sh

PHONY+=angular
angular: angular.tmp ## Upgrade angular

PHONY+=angular-translate
angular-translate: angular-translate.tmp ## Upgrade angular-translate

PHONY+=angular-translate-loader
angular-translate-loader: angular-translate-loader.tmp ## Upgrade angular-translate-loader-static-files

PHONY+=angularUtils
angularUtils: angularUtils.tmp ## Upgrade angularUtils

PHONY+=bootstrap
bootstrap: bootstrap.tmp ## Upgrade bootstrap

PHONY+=fancytree
fancytree: fancytree.tmp ## Upgrade fancytree

PHONY+=jquery
jquery: jquery.tmp ## Upgrade jquery

PHONY+=moment
moment: moment.tmp ## Upgrade moment

.PHONY: $(PHONY)
