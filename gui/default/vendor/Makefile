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

# NVM_VER ?= v0.26.1
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

export PHANTOMJS_SKIP_RECOMMENDED := true 

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

.config_git:
	git config --global url."https://github.com/".insteadOf git://github.com/
	touch $@

PHONY+=preflight
preflight: .config_git $(NVM_DIR)/nvm.sh

# $1 dir
# $2 url
# $3 tag
define clone
	git clone -q $2 $1
	cd $1 && \
		git checkout -q $3
endef

angular.tmp: preflight
	$(call clone,$@,https://github.com/angular/angular.js,$(ANGULAR_VER))
	test -n "$(NVM_DIR)"
	@echo PHANTOMJS_SKIP_RECOMMENDED=$$PHANTOMJS_SKIP_RECOMMENDED
	-grep -q save-exact ~/.npmrc || echo save-exact=true >>~/.npmrc
	# pin all package versions, remove:
	#   "preinstall": "node scripts/npm/check-node-modules.js --purge",
	# which fails the build
	cd $@ && \
	    sed -i 's/"~/"/; s/"^/"/; /preinstall/d;' package.json
	cp -pv package.json $@
	printf '{\n  "directory": "bower_components",\n  "analytics": false,\n  "interactive": false\n}\n' >$@/.bowerrc
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm install --default $(ANGULAR_NODE_VER)
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm use $(ANGULAR_NODE_VER)
	-rm -f ~/.nvmrc
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		nvm ls
	# Fixes: npm ERR! cb() never called!
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install -g npm@$(NPM_VER)
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm version
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm set progress=false
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm set spin=false
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm set loglevel=warn
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install --no-package-lock --no-bin-links
	# this *OVERWRITES* the npm-shrinkwrap.json file, it doesn't APPLY it:
	# ls -l $@/npm-shrinkwrap.json
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm shrinkwrap
	# ls -l $@/npm-shrinkwrap.json
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm dedupe
	# ls -l $@/npm-shrinkwrap.json
	# installs ./node_modules/.bin/bower
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install bower@$(BOWER_VER) # --save-dev
	# version required:
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt@$(GRUNT_VER) # --save-dev
	# installs ./node_modules/.bin/grunt
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install grunt-cli@$(GRUNT_CLI_VER) # --save-dev
	# cd $@ && . $(NVM_DIR)/nvm.sh && \
	# 	npm shrinkwrap
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/bower install
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		npm install --no-package-lock --no-bin-links
	cd $@ && . $(NVM_DIR)/nvm.sh && \
		./node_modules/.bin/grunt package --force
	cd $@ && \
		cp -pv build/angular.js          ../angular && \
		cp -pv build/angular-sanitize.js ../angular

# npm install --no-package-lock --no-bin-links
# npm shrinkwrap
# npm dedupe
# npm install bower --save-dev
# npm install grunt@0.4.2 --save-dev
# npm install grunt-cli@0.1.13 --save-dev
# npm shrinkwrap
# ./node_modules/.bin/bower install
# ./node_modules/.bin/grunt package --force

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

https://github.com/syncthing/syncthing/blob/main/gui/default/vendor/fancytree/jquery.fancytree-all-deps.js

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
all: lfs download asis clean ## Run `make lfs download asis clean`

PHONY+=lfs
lfs: ## Set git to checkout all text files with LF line endings (Windows only)
ifeq ($(OS),Windows_NT)
	sed -i -e 's|^vendor.*|vendor/** eol=lf|' ../../../.gitattributes
endif

PHONY+=asis
asis: ## Set git to checkout all text files "as is" (Windows only)
ifeq ($(OS),Windows_NT)
	git restore ../../../.gitattributes
endif

PHONY+=download
download: $(DIRS) ## Download all the files

PHONY+=clean
clean: ## Clean only the .tmp directories
	find -maxdepth 1 -type d -name "*.tmp" | xargs rm -fr

fmt=\033[36m%-10s\033[0m %s\n
PHONY+=help
help: ## Display a help screen
	@printf '$(fmt)' 'Target' 'Description'
	@printf '$(fmt)' '------' '------------------------------------------'
	@awk -F ':.*##[ \t]*' '/^[^#: \t]+:.*##/ {printf "$(fmt)", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: $(PHONY)
