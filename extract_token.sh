#!/bin/bash

# Nutanix Token Extraction Script
# This script extracts the authentication token from the Nutanix API response

# Configuration
NUTANIX_HOST="10.33.0.88"
NUTANIX_PORT="9440"
USERNAME="admin"
PASSWORD="Nutanix.123"
API_ENDPOINT="/api/nutanix/v3/versions"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔐 Nutanix Token Extraction Script${NC}"
echo -e "${BLUE}================================${NC}"
echo ""

# Function to extract token from set-cookie header
extract_token() {
    local response="$1"
    
    # Extract both cookies and combine them
    local igw_session=$(echo "$response" | grep -o 'NTNX_IGW_SESSION=[^;]*' | sed 's/NTNX_IGW_SESSION=//')
    local mercury_session=$(echo "$response" | grep -o 'NTNX_MERCURY_IGW_SESSION=[^;]*' | sed 's/NTNX_MERCURY_IGW_SESSION=//')
    
    if [ -n "$igw_session" ] && [ -n "$mercury_session" ]; then
        # Combine both cookies in the correct format
        local combined_token="NTNX_IGW_SESSION=${igw_session};NTNX_MERCURY_IGW_SESSION=${mercury_session}"
        
        echo -e "${GREEN}✅ Both cookies extracted successfully!${NC}"
        echo ""
        echo -e "${YELLOW}🔑 Extracted Combined Token:${NC}"
        echo "$combined_token"
        echo ""
        
        # Save token to file
        echo "$combined_token" > nutanix_token.txt
        echo -e "${GREEN}💾 Combined token saved to: nutanix_token.txt${NC}"
        
        # Display token info
        echo ""
        echo -e "${BLUE}📋 Token Information:${NC}"
        echo "Combined Length: ${#combined_token} characters"
        echo "IGW Session Length: ${#igw_session} characters"
        echo "Mercury Session Length: ${#mercury_session} characters"
        
        # Check if IGW session is a JWT token (starts with eyJ)
        if [[ "$igw_session" == eyJ* ]]; then
            echo "IGW Session Type: JWT Token"
            
            # Decode JWT header (first part)
            local header=$(echo "$igw_session" | cut -d'.' -f1 | base64 -d 2>/dev/null)
            if [ $? -eq 0 ]; then
                echo "Algorithm: $(echo "$header" | grep -o '"alg":"[^"]*"' | cut -d'"' -f4)"
                echo "Token Type: $(echo "$header" | grep -o '"typ":"[^"]*"' | cut -d'"' -f4)"
            fi
        else
            echo "IGW Session Type: Custom Token"
        fi
        
        echo "Mercury Session Type: Custom Token"
        
    else
        echo -e "${RED}❌ Failed to extract both cookies from response${NC}"
        echo ""
        echo -e "${YELLOW}🔍 Debug: Found cookies:${NC}"
        echo "IGW Session: $igw_session"
        echo "Mercury Session: $mercury_session"
        echo ""
        echo -e "${YELLOW}🔍 Debug: Full response headers:${NC}"
        echo "$response"
        return 1
    fi
}

# Function to make the API request and extract token
get_token() {
    echo -e "${BLUE}🌐 Making request to Nutanix API...${NC}"
    echo "URL: https://${NUTANIX_HOST}:${NUTANIX_PORT}${API_ENDPOINT}"
    echo "Username: $USERNAME"
    echo ""
    
    # Make the curl request and capture only the headers
    local response=$(curl -skv "https://${NUTANIX_HOST}:${NUTANIX_PORT}${API_ENDPOINT}" \
        -u "${USERNAME}:${PASSWORD}" \
        -D - \
        -o /dev/null 2>&1)
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ API request successful${NC}"
        echo ""
        extract_token "$response"
    else
        echo -e "${RED}❌ API request failed${NC}"
        echo "Error: $response"
        return 1
    fi
}

# Function to validate token format
validate_token() {
    local token="$1"
    
    echo ""
    echo -e "${BLUE}🔍 Validating combined token format...${NC}"
    
    # Check if token is not empty
    if [ -z "$token" ]; then
        echo -e "${RED}❌ Token is empty${NC}"
        return 1
    fi
    
    # Check if token contains both required cookies
    if [[ "$token" == *"NTNX_IGW_SESSION="* ]] && [[ "$token" == *"NTNX_MERCURY_IGW_SESSION="* ]]; then
        echo -e "${GREEN}✅ Token contains both required cookies${NC}"
    else
        echo -e "${RED}❌ Token missing required cookies${NC}"
        return 1
    fi
    
    # Check if token has correct separator
    if [[ "$token" == *";"* ]]; then
        echo -e "${GREEN}✅ Token has correct cookie separator (;)${NC}"
    else
        echo -e "${RED}❌ Token missing cookie separator${NC}"
        return 1
    fi
    
    # Extract individual cookies for validation
    local igw_session=$(echo "$token" | sed 's/.*NTNX_IGW_SESSION=\([^;]*\).*/\1/')
    local mercury_session=$(echo "$token" | sed 's/.*NTNX_MERCURY_IGW_SESSION=\([^;]*\).*/\1/')
    
    # Validate IGW session (should be JWT)
    if [[ "$igw_session" == eyJ* ]]; then
        echo -e "${GREEN}✅ IGW Session is valid JWT format${NC}"
        
        # Check JWT structure (3 parts separated by dots)
        local parts=$(echo "$igw_session" | tr -cd '.' | wc -c)
        if [ "$parts" -eq 2 ]; then
            echo -e "${GREEN}✅ IGW Session has valid JWT structure (3 parts)${NC}"
        else
            echo -e "${YELLOW}⚠️  IGW Session doesn't have standard JWT structure${NC}"
        fi
    else
        echo -e "${YELLOW}⚠️  IGW Session is not in JWT format${NC}"
    fi
    
    # Validate Mercury session
    if [ -n "$mercury_session" ]; then
        echo -e "${GREEN}✅ Mercury Session is present${NC}"
    else
        echo -e "${RED}❌ Mercury Session is missing${NC}"
        return 1
    fi
    
    # Check combined token length (should be reasonable)
    if [ ${#token} -gt 200 ]; then
        echo -e "${GREEN}✅ Combined token length is reasonable (${#token} chars)${NC}"
    else
        echo -e "${YELLOW}⚠️  Combined token seems short (${#token} chars)${NC}"
    fi
    
    return 0
}

# Main execution
main() {
    # Check if curl is available
    if ! command -v curl &> /dev/null; then
        echo -e "${RED}❌ Error: curl is not installed${NC}"
        exit 1
    fi
    
    # Check if base64 is available (for JWT decoding)
    if ! command -v base64 &> /dev/null; then
        echo -e "${YELLOW}⚠️  Warning: base64 not available, JWT decoding disabled${NC}"
    fi
    
    # Get the token
    if get_token; then
        # Read the saved token and validate it
        if [ -f "nutanix_token.txt" ]; then
            local saved_token=$(cat nutanix_token.txt)
            validate_token "$saved_token"
        fi
        
        echo ""
        echo -e "${GREEN}🎉 Token extraction completed successfully!${NC}"
        echo ""
        echo -e "${BLUE}📖 Usage:${NC}"
        echo "The combined token has been saved to 'nutanix_token.txt'"
        echo "You can use it in subsequent API calls like:"
        echo "curl -H 'Cookie: $(cat nutanix_token.txt)' https://${NUTANIX_HOST}:${NUTANIX_PORT}/api/..."
        echo ""
        echo "Or use it directly:"
        echo "TOKEN=\$(cat nutanix_token.txt)"
        echo "curl -H \"Cookie: \$TOKEN\" https://${NUTANIX_HOST}:${NUTANIX_PORT}/api/..."
        
    else
        echo -e "${RED}❌ Token extraction failed${NC}"
        exit 1
    fi
}

# Run the main function
main "$@"
