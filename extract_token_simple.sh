#!/bin/bash

# Simple Nutanix Token Extraction - Extracts both cookies and combines them
# Usage: ./extract_token_simple.sh
# Sets NUTANIX_COOKIE environment variable for use in commands

# Extract both cookies and combine them in the correct format
COOKIE_VALUE=$(curl -skv "https://10.33.0.88:9440/api/nutanix/v3/versions" \
  -u "admin:Nutanix.123" \
  -D - \
  -o /dev/null 2>&1 | \
  awk '
    /< set-cookie: NTNX_IGW_SESSION=/ {
      gsub(/< set-cookie: NTNX_IGW_SESSION=/, "")
      gsub(/;Expires=.*/, "")
      igw_session = $0
    }
    /< set-cookie: NTNX_MERCURY_IGW_SESSION=/ {
      gsub(/< set-cookie: NTNX_MERCURY_IGW_SESSION=/, "")
      gsub(/;Expires=.*/, "")
      mercury_session = $0
    }
    END {
      if (igw_session && mercury_session) {
        print "NTNX_IGW_SESSION=" igw_session ";NTNX_MERCURY_IGW_SESSION=" mercury_session
      }
    }
  ')

# Save to file
echo "$COOKIE_VALUE" > nutanix_token.txt

# Set environment variable
export NUTANIX_COOKIE="$COOKIE_VALUE"

