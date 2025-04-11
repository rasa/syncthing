#!/usr/bin/make

PHONY+=default
default: help ;

ANGULAR_VER ?= v1.3.20
ANGULAR_TRANSLATE_VER ?= 2.9.0
ANGULAR_TRANSLATE_LOADER_STATIC_FILES_VER ?= 2.11.0
ANGULAR_UTILS_VER ?= 759009c
BOOTSTRAP_VER ?= v3.3.6
FANCYTREE_VER ?= v2.38.0
JQUERY_VER ?= 2.2.2
MOMENT_VER ?= 2.19.4

NVM_VER ?= v0.40.2
NPM_VER ?= 2.15.11

# angular:
ANGULAR_NODE_VER ?= 0.10.48
BOWER_VER ?= 1.3.12
GRUNT_VER ?= 0.4.2
GRUNT_CLI_VER ?= 0.1.13

# angular-translate:
NODE_VER ?= 4.2

export NVM_DIR := $(shell [ -n "$${XDG_CONFIG_HOME}" ] && printf %s "$${XDG_CONFIG_HOME}/nvm" || printf %s "$${HOME}/.nvm")

export PYTHON := $(shell which python2.7)

ifeq (, $(PYTHON))
	$(error "Cannot find python2.7")
	$(error "To install on Ubuntu 23.10+, run: make -f python2.mak install")
endif

DIRS = angular.tmp \
	angular-translate.tmp \
	angular-translate-loader-static-files.tmp \
	angularUtils.tmp \
	bootstrap.tmp \
	fancytree.tmp \
	jquery.tmp \
	moment.tmp

$(NVM_DIR)/nvm.sh:
	test -f "$@" || \
		curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/$(NVM_VER)/install.sh | bash

PHONY+=nvm
nvm: $(NVM_DIR)/nvm.sh

# GitHub doesn't support git:// URLs any more
# .git_configured:
# 	# git config --global url."https://github.com/".insteadOf git://github.com/
# 	touch $@

# PHONY+=preflight
# preflight: nvm .git_configured

# $1 dir
# $2 url
# $3 tag
define clone
	git clone -q $2 $1
	cd $1 && \
		git checkout -q $3
endef

angular.tmp: $(NVM_DIR)/nvm.sh
	$(call clone,$@,https://github.com/angular/angular.js,$(ANGULAR_VER))
	# change:
	#   "stringmap": "0.2.2"
	# to:
	#   "stringmap": "0.2.2","request": "2.51.0", "tough-cookie": "0.12.1"
	cp -pv package.json $@
	#   "preinstall": "node scripts/npm/check-node-modules.js --purge",
	# which fails the build
	cd $@ && \
	    sed -i '/preinstall/d;' package.json
	# pin all package versions
	# change git: URLs to https: as GitHub doesn't support git: URLs any more
	cd $@ && \
	    sed -i 's/"~/"/; s/"^/"/; s|git://|https://|;' package.json npm-shrinkwrap.json npm-shrinkwrap.clean.json
	printf '{\n  "directory": "bower_components",\n  "analytics": false,\n  "interactive": false\n}\n' >$@/.bowerrc
	# -rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm install --default $(ANGULAR_NODE_VER)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm use $(ANGULAR_NODE_VER)
	# fixes: npm ERR! cb() never called!
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g npm@$(NPM_VER)
	# Try without: --no-package-lock --no-bin-links
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install --no-package-lock --no-bin-links
	# installs ./node_modules/.bin/bower
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install bower@$(BOWER_VER)
	# version required:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt@$(GRUNT_VER)
	# installs ./node_modules/.bin/grunt
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt-cli@$(GRUNT_CLI_VER)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/bower install
	# Try without: --no-package-lock --no-bin-links
	# Try without running npm install again
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install --no-package-lock --no-bin-links
	# Try without: --force
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/grunt package --force
	cd $@ && \
		cp -pv build/angular.js          ../angular && \
		cp -pv build/angular-sanitize.js ../angular

angular-translate.tmp: angular.tmp
	$(call clone,$@,https://github.com/angular-translate/angular-translate,$(ANGULAR_TRANSLATE_VER))
	cd $@ && \
	    sed -i "/karma-phantomjs-launcher/d" package.json
	cd $@ && \
	    sed -i 's/"~/"/; s/"^/"/;' package.json
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm uninstall $(ANGULAR_NODE_VER)
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm install --default $(NODE_VER)
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm use $(NODE_VER)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm ls
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install bower
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt@$(GRUNT_VER) --save-dev
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt-cli@$(GRUNT_CLI_VER) --save-dev
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install
	printf '{\n  "directory": "bower_components",\n  "analytics": false,\n  "interactive": false\n}\n' >$@/.bowerrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/bower install
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/grunt package --force
	cd $@ && \
		cp -pv build/angular-translate.js ../angular

angular-translate-loader-static-files.tmp: angular.tmp
	$(call clone,$@,https://github.com/angular-translate/bower-angular-translate-loader-static-files,$(ANGULAR_TRANSLATE_LOADER_STATIC_FILES_VER))
	cd $@ && \
		cp -pv angular-translate-loader-static-files.js ../angular

angularUtils.tmp: angular.tmp
	$(call clone,$@,https://github.com/michaelbromley/angularUtils,$(ANGULAR_UTILS_VER)
	cd $@ && \
		cp -pv src/directives/pagination/dirPagination.js ../angular/angular-dirPagination.js

bootstrap.tmp:
	$(call clone,$@,https://github.com/twbs/bootstrap,$(BOOTSTRAP_VER))
	cd $@ && \
		cp -pv dist/css/bootstrap.css       ../bootstrap/css && \
		cp -pv dist/css/bootstrap-theme.css ../bootstrap/css && \
		cp -pv dist/js/bootstrap.js         ../bootstrap/js && \
		cp -pv LICENSE                      ../bootstrap/LICENSE

fancytree.tmp:
	$(call clone,$@,https://github.com/mar10/fancytree,$(FANCYTREE_VER))
	cd $@ && \
		cp -pv dist/jquery.fancytree-all-deps.js ../fancytree && \
		cp -pv dist/dist/LICENSE.txt ../fancytree

jquery.tmp:
	$(call clone,$@,https://github.com/jquery/jquery,$(JQUERY_VER))
	cd $@ && \
		cp -pv dist/jquery.js ../jquery/jquery-$(JQUERY_VER).js && \
		cp -pv LICENSE.txt    ../jquery/LICENSE

moment.tmp:
	$(call clone,$@,https://github.com/moment/moment,$(MOMENT_VER))
	cd $@ && \
		cp -pv moment.js ../moment
		cp -pv LICENSE   ../moment

PHONY+=all
all: checkout_lfs download checkout_asis clean ## Run `make checkout_lfs download checkout_asis clean`

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

fmt=\033[36m%-13s\033[0m %s\n
PHONY+=help
help: ## Display a help screen
	@printf '$(fmt)' 'Target' 'Description'
	@printf '$(fmt)' '------' '------------------------------------------'
	@awk -F ':.*##[ \t]*' '/^[^#: \t]+:.*##/ {printf "$(fmt)", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: $(PHONY)
