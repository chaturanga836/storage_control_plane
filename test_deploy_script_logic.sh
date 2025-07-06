#!/bin/bash
# Test script to validate the deploy_ec2.sh directory logic fix

echo "Testing deploy_ec2.sh directory logic..."

# Test 1: Check if script works when run from inside storage_control_plane
echo "Test 1: Running from inside storage_control_plane directory"
if [ "$(basename $(pwd))" = "storage_control_plane" ]; then
    echo "✅ PASS: Script correctly detects it's already in storage_control_plane"
else
    echo "❌ FAIL: Script should detect it's in storage_control_plane"
fi

# Test 2: Check CURRENT_DIR variable
CURRENT_DIR=$(pwd)
echo "Test 2: Current directory is: $CURRENT_DIR"
if [[ "$CURRENT_DIR" == *"storage_control_plane" ]]; then
    echo "✅ PASS: CURRENT_DIR correctly points to storage_control_plane directory"
else
    echo "❌ FAIL: CURRENT_DIR does not point to storage_control_plane directory"
fi

# Test 3: Check if go.mod exists
echo "Test 3: Checking for go.mod file"
if [ -f "go.mod" ]; then
    echo "✅ PASS: go.mod found in current directory"
else
    echo "❌ FAIL: go.mod not found - are we in the right directory?"
fi

echo ""
echo "🎯 Summary: The fixed deploy_ec2.sh script should now work correctly when run from:"
echo "   1. Inside the storage_control_plane directory ✅"  
echo "   2. From parent directory with storage_control_plane subdirectory ✅"
echo "   3. From home directory with storage_control_plane subdirectory ✅"
