#!/bin/bash

# Script to set up GitHub token for go-sample-app

echo "GitHub Token Setup for go-sample-app"
echo "===================================="
echo ""

# Check if GITHUB_TOKEN is already set
if [ -n "$GITHUB_TOKEN" ]; then
    echo "✅ GITHUB_TOKEN is already set in your environment"
    echo "Token: ${GITHUB_TOKEN:0:10}..."
    echo ""
    echo "You can now run:"
    echo "export CONFIG_REMOTE_URL='https://github.com/KyberNetwork/af_multi_lang_poc/blob/main/go-sample-app/config/config.yaml'"
    echo "go run main.go -config-type=remote"
    exit 0
fi

echo "To access private GitHub repositories, you need a Personal Access Token."
echo ""
echo "1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)"
echo "2. Click 'Generate new token (classic)'"
echo "3. Give it a name like 'go-sample-app-config'"
echo "4. Select scopes: 'repo' (for private repositories)"
echo "5. Click 'Generate token'"
echo "6. Copy the token (you won't see it again!)"
echo ""

read -p "Enter your GitHub Personal Access Token: " token

if [ -z "$token" ]; then
    echo "❌ No token provided. Exiting."
    exit 1
fi

# Export the token
export GITHUB_TOKEN="$token"

echo ""
echo "✅ GitHub token exported as GITHUB_TOKEN"
echo ""

# Test the token
echo "Testing token access..."
if curl -s -H "Authorization: token $token" -H "User-Agent: go-sample-app/1.0" \
   "https://api.github.com/user" | grep -q "login"; then
    echo "✅ Token is valid!"
else
    echo "❌ Token validation failed. Please check your token."
    exit 1
fi

echo ""
echo "Now you can run the application with:"
echo "export CONFIG_REMOTE_URL='https://github.com/KyberNetwork/af_multi_lang_poc/blob/main/go-sample-app/config/config.yaml'"
echo "go run main.go -config-type=remote"
echo ""
echo "Or run this script again to set the token for future sessions." 