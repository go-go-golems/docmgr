#!/bin/bash

API_URL="http://localhost:8080"

echo "=== Testing docmgr API ==="
echo

echo "1. List all workspaces:"
curl -s "$API_URL/api/list" | python3 -m json.tool
echo
echo

echo "2. Get documents for MEN-3475:"
curl -s "$API_URL/api/documents?ticket=MEN-3475" | python3 -m json.tool | head -30
echo
echo

echo "3. Search for 'architecture':"
curl -s "$API_URL/api/search?q=architecture" | python3 -m json.tool | head -30
echo
echo

echo "4. Filter by topic 'chat':"
curl -s "$API_URL/api/search?topic=chat" | python3 -m json.tool | head -30
echo
echo

echo "5. Filter by document type 'reference':"
curl -s "$API_URL/api/search?type=reference" | python3 -m json.tool | head -30
echo
echo

echo "6. Combined search - topic 'chat' + type 'design':"
curl -s "$API_URL/api/search?topic=chat&type=design" | python3 -m json.tool | head -30
echo

echo "=== Tests complete ==="
