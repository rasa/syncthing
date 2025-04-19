#!/usr/bin/make

SHELL := bash

PHONY+=default
default: help ;

# What syncthing is using now:
ANGULAR_VER ?= v1.3.20
ANGULAR_TRANSLATE_VER ?= 2.9.0
ANGULAR_TRANSLATE_LOADER_VER ?= 2.11.0
ANGULAR_UTILS_VER ?= 759009c
# It's v3.3.6, per
# https://github.com/syncthing/syncthing/commit/c6881b6d02b31ef00190ae0de9e2fba5cec8a7aa
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
A_NODE_VER ?= 0.10.48
REQUEST_VER ?= 2.51.0

# angular-translate 2.9.0's versions:
B_NODE_VER ?= 4.2.6
GRUNT_EXPRESS_VER ?= 0.2.1
GRUNT_KARMA_VER ?= 0.8.3
KARMA_VER ?= 0.12.31

ifeq (, $(XDG_CONFIG_HOME))
	$(error "XDG_CONFIG_HOME is not defined")
endif

NVM_DIR := $(shell printf %s "$${XDG_CONFIG_HOME}/nvm")

PYTHON := $(shell which python2.7)

ifeq (, $(PYTHON))
	$(error "Cannot find python2.7")
	$(error "To install on Ubuntu 23.10+, run: make -f python2.mak install")
endif

DIRS = \
	angular.tmp \
	angular-translate.tmp \
	angular-translate-loader.tmp \
	angularUtils.tmp \
	bootstrap.tmp \
	fancytree.tmp \
	jquery.tmp \
	moment.tmp

# $1 url
# $2 tag
define init
	test -d $@ || git clone -q $1 $@
	@: # git reset --hard
	@: # git clean -fdx
	cd $@ && \
		git checkout -q $2
endef

# $1 grunt command
define build
	@: # Stop bower from prompting:
	@: #  [?] May bower anonymously report usage statistics to improve the tool over time? (Y/n)
	printf '{"interactive": false}' >$@/.bowerrc
	@:
	@: # Change git: URLs to https: (Github dropped support for git: URLs in 2022):
	find $@ -type f -name 'bower.json' -o -name package.json -o -name 'npm-shrinkwrap*.json' | xargs \
		sed -E -i -f '$(FIXJSONS_SED)'
	@:
	@: # Fix ES6 error: npm ERR! cb() never called!
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g npm@$(NPM_VER) --silent
	@:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install \
			grunt@$(GRUNT_VER) \
			grunt-express@$(GRUNT_EXPRESS_VER) \
			grunt-karma@$(GRUNT_KARMA_VER) \
			karma@$(KARMA_VER) \
	 		request@$(REQUEST_VER) \
	 		tough-cookie@$(TOUCH_COOKIE_VER) \
			--silent
	@:
	@: # probably unneeded, but it can't hurt
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm cache clean
	@:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install --silent
	@:
	@: # probably unneeded, but it can't hurt
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm dedup
	@:
	@: # needed, since shared installs everything
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm prune
	@:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm shrinkwrap --dev
	@:
	@: # without the version number, 1.8.3 would be installed
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g bower@$(BOWER_VER) --silent
	@:
	@: # without the version number, 1.5.0 would be installed
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt-cli@$(GRUNT_CLI_VER) --silent
	@:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		PATH="./node_modules/.bin/:$${PATH}" \
			bower install --quiet
	@:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		PATH="./node_modules/.bin/:$${PATH}" \
			grunt $1 --quiet
endef

# $1 dir.tmp
# $2 dest_dir
define fini
	cd "$@" && \
		git for-each-ref --sort=creatordate --format '%(creatordate:short) %(refname:short)' refs/tags > "../$(call basename,$@)-tags.txt"
	cd "$1" && \
		git diff -uw . >"../$(call basename,$@).diff"
	cd "$1" && \
		git stash push -m "$(call basename,$@)" .
	-touch $2
endef

$(NVM_DIR)/nvm.sh:
	test -f "$@" || \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/$(NVM_VER)/install.sh | bash

PHONY+=install_nvm
install_nvm: $(NVM_DIR)/nvm.sh ## Install the latest nvm

# GitHub stopped supporting git:// URLs:
.git_urls.fixed:
	git config --global url."https://github.com/".insteadOf git://github.com/
	touch $@

FIXJSONS_SED := .fixjsons.sed

$(FIXJSONS_SED):
	@: # Change git: URLs to https: as GitHub stopped supporting git:// URLs:
	printf '%s\n' 's|git://|https://|;' >$@
	@: # Pin exact versions, not semver ranges:
	printf '%s\n' 's/"~/"/; s/"^/"/;' >>$@
	@: # Stop pulling in ES6 code:
	printf '%s\n' '/karma-phantomjs-launcher/d;' >>$@
	@: # Pin angular version (not required, but it can't hurt):
	printf '%s\n' 's/("angular":\s*)"[^"]+"(.*)/\1"$(ANGULAR_VER)"\2/;' >>$@
	@: # Pin all package versions:
	printf '%s\n' 's/("grunt":\s*)"[^"]+"(.*)/\1"$(GRUNT_VER)"\2/;' >>$@
	printf '%s\n' 's/("grunt-cli":\s*)"[^"]+"(.*)/\1"$(GRUNT_CLI_VER)"\2/;' >>$@
	printf '%s\n' 's/("grunt-express":\s*)"[^"]+"(.*)/\1"$(GRUNT_EXPRESS_VER)"\2/;' >>$@
	printf '%s\n' 's/("grunt-karma":\s*)"[^"]+"(.*)/\1"$(GRUNT_KARMA_VER)"\2/;' >>$@
	printf '%s\n' 's/("karma":\s*)"[^"]+"(.*)/\1"$(KARMA_VER)"\2/;' >>$@
	printf '%s\n' 's/("tough-cookie":\s*)"[^"]"+(.*)/\1"$(TOUCH_COOKIE_VER)"\2/;' >>$@

PHONY+=setup
setup: $(NVM_DIR)/nvm.sh .git_urls.fixed $(FIXJSONS_SED) ## install the latest nvm, configure git to rewrite git:// URLs

PHONY+=install_node
install_node: install_nvm ## Install node using the NODE_VER variable
	test -n "$(NODE_VER)"
	. $(NVM_DIR)/nvm.sh && \
			nvm install --default $(NODE_VER)
	. $(NVM_DIR)/nvm.sh && \
		nvm alias default $(NODE_VER)

angular.tmp: setup
	$(MAKE) install_node NODE_VER="$(A_NODE_VER)"
	$(call init,https://github.com/angular/angular.js,$(ANGULAR_VER))
	$(call build,package)
	cp -pv $@/build/angular.js          angular
	cp -pv $@/build/angular-sanitize.js angular
	-cp -pv $@/LICENSE angular/angular.js.LICENSE
	-cp -pv $@/LICENSE angular/angular-sanitize.js.LICENSE
	$(call fini,angular)

angular-translate.tmp: setup
	$(MAKE) install_node NODE_VER="$(B_NODE_VER)"
	$(call init,https://github.com/angular-translate/angular-translate,$(ANGULAR_TRANSLATE_VER))
	$(call build,build:core)
	cp -pv $@/dist/angular-translate.js angular
	-cp -pv $@/LICENSE angular/angular-translate.js.LICENSE
	$(call fini,angular)

angular-translate-loader.tmp:
	$(call init,https://github.com/angular-translate/bower-angular-translate-loader-static-files,$(ANGULAR_TRANSLATE_LOADER_VER))
	cp -pv $@/angular-translate-loader-static-files.js angular
	@: # see https://github.com/angular-translate/bower-angular-translate-loader-static-files?tab=readme-ov-file#license
	-cp -pv angular/angular-translate.js.LICENSE angular/angular-translate-loader-static-files.js.LICENSE
	$(call fini,angular)

angularUtils.tmp:
	$(call init,https://github.com/michaelbromley/angularUtils,$(ANGULAR_UTILS_VER))
	cp -pv $@/src/directives/pagination/dirPagination.js angular/angular-dirPagination.js
	-cp -pv $@/license.txt angular/angular-dirPagination.js.LICENSE
	$(call fini,angular)

bootstrap.tmp:
	$(call init,https://github.com/twbs/bootstrap,$(BOOTSTRAP_VER))
	cp -pv $@/dist/css/bootstrap.css       bootstrap/css
	cp -pv $@/dist/css/bootstrap-theme.css bootstrap/css
	cp -pv $@/dist/js/bootstrap.js         bootstrap/js
	-cp -pv $@/LICENSE                      bootstrap/LICENSE
	$(call fini,bootstrap)

fancytree.tmp:
	$(call init,https://github.com/mar10/fancytree,$(FANCYTREE_VER))
	cd $@/dist && \
		patch --batch --ignore-whitespace --strip=5 < ../../8509.patch
	cd $@/dist && \
		patch --batch --ignore-whitespace  --strip=1 < ../../f085361.patch
	cp -pv $@/dist/jquery.fancytree-all-deps.js fancytree
	-cp -pv $@/dist/LICENSE.txt fancytree
	$(call fini,fancytree)

jquery.tmp:
	$(call init,https://github.com/jquery/jquery,$(JQUERY_VER))
	cp -pv $@/dist/jquery.js jquery/jquery-$(JQUERY_VER).js
	-cp -pv $@/LICENSE.txt    jquery/LICENSE
	$(call fini,jquery)

moment.tmp:
	$(call init,https://github.com/moment/moment,$(MOMENT_VER))
	cp -pv $@/moment.js moment
	-cp -pv $@/LICENSE   moment
	$(call fini,moment)

PHONY+=all
all: checkout_lfs build checkout_asis ## Run `make checkout_lfs build checkout_asis`

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

PHONY+=build
build: $(DIRS) ## Build all the vendored repositories

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

PHONY+=clean
clean: ## Clean just the .tmp directories, and downloaded .deb files
	find -maxdepth 1 -type d -name "*.tmp" | xargs rm -fr
	rm -fr *.deb

PHONY+=nuke
nuke: clean ## Clean all downloaded/generated files
	-rm -fr *-tags.txt
	-rm -fr *.diff
	-rm -fr $(FIXJSONS_SED)

PHONY+=uninstall_all
uninstall_all: ## Uninstall nvm and all node versions by wiping NVM_DIR, ~/.npm, etc.
	test -n "$(NVM_DIR)"
	-sed -i '/nvm/d; /NVM/d;' "$${HOME}/.bashrc"
	-find \
		"$(NVM_DIR)" \
		"$${HOME}/.config/configstore" \
		"$${HOME}/.node-gyp" \
		"$${HOME}/.npm" \
		"$${HOME}/.nvmrc" \
		-mindepth 1 -depth -exec rm -rfv -- {} + 2>/dev/null

fmt=\033[36m%-24s\033[0m %s\n
PHONY+=help
help: ## Display a help screen
	@printf '$(fmt)' 'Target' 'Description'
	@printf '$(fmt)' '------' '------------------------------------------'
	@awk -F ':.*##[ \t]*' '/^[^#: \t]+:.*##/ {printf "$(fmt)", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: $(PHONY)
