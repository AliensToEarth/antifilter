# Makefile converted from run.sh
# Handles geoip and geosite data processing

# Variables
YEAR := $(shell date +%Y)
MONTH := $(shell date +%m)
DBIP_URL := https://download.db-ip.com/free/dbip-country-lite-$(YEAR)-$(MONTH).mmdb.gz
GEOIP_DIR := geoip
GEOSITE_DIR := geosite
GEOSITE_PARSER_DIR := geosite-parser
GEOIP_CONFIG := geoip-antifilter.json
GEOSITE_CONFIG := geosite-antifilter.json
DBIP_FILE := dbip-country-lite.mmdb.gz
DBIP_EXTRACTED := dbip-country-lite.mmdb
ANTIFILTER_FILE := antifilter-comunity

# Phony targets
.PHONY: all geoip geosite clean help submodule-init submodule-update
.PHONY: geoip-download geoip-extract geoip-build
.PHONY: geosite-download geosite-build geosite-parse

# Default target
all: submodule-init geoip geosite

# Main geoip workflow
geoip: submodule-init geoip-download geoip-extract geoip-build

# Main geosite workflow
geosite: submodule-init geosite-download geosite-build geosite-parse

# Submodule initialization
submodule-init:
	@echo "Initializing and updating git submodules..."
	git submodule update --init --recursive

# Submodule update with change detection
submodule-update:
	@echo "Checking for submodule updates..."
	@echo "Current submodule commit hashes:"
	@git submodule status
	@echo "Fetching latest changes from remote repositories..."
	git submodule foreach 'git fetch origin'
	@echo "Checking for updates..."
	@if git submodule status | grep -q '^+'; then \
		echo "Updates available, updating submodules..."; \
		git submodule update --remote --recursive; \
		echo "Updated submodule commit hashes:"; \
		git submodule status; \
	else \
		echo "No updates available, submodules are up to date."; \
	fi

# GeoIP targets
geoip-download:
	@echo "Downloading DB-IP database for $(YEAR)-$(MONTH)..."
	cd $(GEOIP_DIR) && curl -L -o $(DBIP_FILE) $(DBIP_URL)

geoip-extract: geoip-download
	@echo "Extracting and organizing DB-IP database..."
	cd $(GEOIP_DIR) && gzip -d $(DBIP_FILE)
	cd $(GEOIP_DIR) && mkdir -p db-ip
	cd $(GEOIP_DIR) && mv dbip-country-lite*.mmdb ./db-ip/$(DBIP_EXTRACTED)

geoip-build: geoip-extract
	@echo "Building geoip data..."
	cd $(GEOIP_DIR) && go run ./ -c ../$(GEOIP_CONFIG)

# GeoSite targets
geosite-download:
	@echo "Downloading antifilter community domains..."
	cd $(GEOSITE_DIR) && curl -L "https://community.antifilter.download/list/domains.lst" -o ./data/$(ANTIFILTER_FILE)

geosite-build: geosite-download
	@echo "Building geosite data..."
	cd $(GEOSITE_DIR) && go run ./ --outputdir=../output --outputname="geosite.dat"

geosite-parse: geosite-build
	@echo "Processing geosite data with antifilter configuration..."
	cd $(GEOSITE_PARSER_DIR) && go mod tidy
	cd $(GEOSITE_PARSER_DIR) && go run ./ -c ../$(GEOSITE_CONFIG)

# Utility targets
clean:
	@echo "Cleaning up downloaded and generated files..."
	-rm -f $(GEOIP_DIR)/$(DBIP_FILE)
	-rm -f $(GEOIP_DIR)/dbip-country-lite*.mmdb
	-rm -rf $(GEOIP_DIR)/db-ip
	-rm -f $(GEOSITE_DIR)/data/$(ANTIFILTER_FILE)
	-rm -rf output

help:
	@echo "Available targets:"
	@echo "  all          - Run both geoip and geosite workflows (default)"
	@echo "  geoip        - Run complete geoip workflow"
	@echo "  geosite      - Run complete geosite workflow"
	@echo ""
	@echo "Setup targets:"
	@echo "  submodule-init   - Initialize and update git submodules"
	@echo "  submodule-update - Check and update submodules if changes available"
	@echo ""
	@echo "Individual geoip targets:"
	@echo "  geoip-download - Download DB-IP country database"
	@echo "  geoip-extract  - Extract and organize database file"
	@echo "  geoip-build    - Build geoip data with configuration"
	@echo ""
	@echo "Individual geosite targets:"
	@echo "  geosite-download - Download antifilter community domains"
	@echo "  geosite-build    - Build geosite data with filters"
	@echo "  geosite-parse    - Process geosite data with antifilter configuration"
	@echo ""
	@echo "Utility targets:"
	@echo "  clean        - Remove downloaded and generated files"
	@echo "  help         - Show this help message"
	@echo ""
	@echo "Current settings:"
	@echo "  Year/Month:  $(YEAR)-$(MONTH)"
	@echo "  DB-IP URL:   $(DBIP_URL)"