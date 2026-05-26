# Makefile Usage Guide

This Makefile replaces the functionality of `run.sh` with organized, dependency-aware targets.

## Quick Start

```bash
# Run everything (equivalent to ./run.sh)
make

# Initialize submodules first (recommended for new setup)
make submodule-init

# Check and update submodules if changes are available
make submodule-update

# Run only geoip workflow
make geoip

# Run only geosite workflow
make geosite

# Show all available targets
make help

# Clean up generated files
make clean
```

## Target Overview

### Main Workflows

- **`make all`** (default) - Runs both geoip and geosite workflows
- **`make geoip`** - Complete geoip workflow: submodule-init → download → extract → build
- **`make geosite`** - Complete geosite workflow: submodule-init → download → build (including parser step)

### Setup Targets

- **`make submodule-init`** - Initialize and update git submodules (geoip and geosite)
- **`make submodule-update`** - Check current commit hashes and update submodules if changes are available

### Individual GeoIP Targets

- **`make geoip-download`** - Download DB-IP country database for current month
- **`make geoip-extract`** - Extract and organize the database file
- **`make geoip-build`** - Build geoip data using configuration

### Individual GeoSite Targets

- **`make geosite-download`** - Download antifilter community domains list
- **`make geosite-build`** - Build geosite data and run the parser with the antifilter configuration

### Utility Targets

- **`make clean`** - Remove all downloaded and generated files
- **`make help`** - Display detailed help information

## Key Features

### Dynamic Date Handling
The Makefile automatically uses the current year and month for DB-IP database downloads:
```
Current format: https://download.db-ip.com/free/dbip-country-lite-YYYY-MM.mmdb.gz
```

### Dependency Management
Targets have proper dependencies to ensure correct execution order:
- `submodule-init` runs first to ensure repositories are available
- `geoip-extract` depends on `geoip-download`
- `geoip-build` depends on `geoip-extract`
- Both `geoip` and `geosite` depend on `submodule-init`
- `geosite-build` includes the parser step and runs automatically from `geosite`

### GeoSite Parser Integration
The geosite workflow now includes an additional parsing step that processes the generated `geosite.dat` file using the `geosite-antifilter.json` configuration. This creates filtered output files containing only specified domain lists (google, youtube, meta, antifilter-community, private).

### Error Handling
- Commands use proper error propagation
- Clean target uses `-` prefix to ignore missing files
- Each step provides clear status messages

## Advanced Usage

### Dry Run
Test what commands will be executed without running them:
```bash
make --dry-run geoip
make -n geosite
```

### Parallel Execution
Run geoip and geosite workflows in parallel:
```bash
make -j2 geoip geosite
```

### Verbose Output
See all commands being executed:
```bash
make --debug=v geoip
```

## Troubleshooting

### Common Issues

1. **Missing directories**: Ensure `geoip/` and `geosite/` directories exist
2. **Network issues**: Check internet connection for download targets
3. **Go build errors**: Verify Go is installed and projects compile correctly

### Debug Commands

```bash
# Check Makefile syntax
make --dry-run help

# Verify current date variables
make help | grep "Year/Month"

# Test individual targets
make --dry-run geosite-build
```

## GeoSite Parser

The project includes a custom geosite parser utility located in the `geosite-parser/` directory. This utility:

- Processes the generated `geosite.dat` file
- Uses `geosite-antifilter.json` configuration
- Creates filtered output with specific domain lists
- Runs automatically after `make geosite`

For detailed information about the parser, see [`geosite-parser/README.md`](geosite-parser/README.md).