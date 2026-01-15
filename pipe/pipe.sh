#!/usr/bin/env bash

#
# SafeDep Vet Pipe for policy driven vetting of open source dependencies.
#

echo ""
echo "Executing SafeDep Vet Pipe..."
echo ""

# Set Default values to default variables
# Its not possible to set them in pipe.yml
export VET_VERSION=${VET_VERSION:-"latest"}
export CLOUD=${CLOUD:-"false"}
export TIMEOUT=${TIMEOUT:-300}
export SKIP_FILTER_CI_FAIL=${SKIP_FILTER_CI_FAIL:-"false"}

# Pre-scan Checks
if ! command -v vet &> /dev/null
then
    echo "Error: vet is not installed"
    exit 1
fi

if [ -n "$POLICY" ] && [ ! -f "$POLICY" ]; then
    echo "Policy file not found: $POLICY"
    exit 1
fi

if [ -n "$EXCEPTION_FILE" ] && [ ! -f "$EXCEPTION_FILE" ]; then
    echo "Exception file not found: $EXCEPTION_FILE"
    exit 1
fi

if [ "$CLOUD" = "true" ]; then
    if [ -z "$CLOUD_KEY" ] || [ -z "$CLOUD_TENANT" ]; then
        echo "Cloud key and tenant must be provided when cloud is enabled"
        exit 1
    fi
fi

# Scan Execution
ARTIFACT_FILENAME="vet-report.json"
CMD="vet scan -s --report-json $ARTIFACT_FILENAME"

if [ -n "$POLICY" ]; then
    CMD="$CMD --filter-suite $POLICY"
    if [ "$SKIP_FILTER_CI_FAIL" = "false" ]; then
        CMD="$CMD --filter-fail --fail-fast"
    fi
fi

if [ -n "$EXCEPTION_FILE" ]; then
    CMD="$CMD --exceptions $EXCEPTION_FILE"
fi

if [ "$CLOUD" = "true" ]; then
    CMD="$CMD --report-sync"
    CMD="$CMD --report-sync-project $BITBUCKET_REPO_FULL_NAME"
    CMD="$CMD --report-sync-project-version $BITBUCKET_BRANCH"
    CMD="$CMD --malware"
    CMD="$CMD --malware-analysis-timeout ${TIMEOUT}s"
fi

if [ -n "$TRUSTED_REGISTRIES" ]; then
    IFS=',' read -ra REGISTRIES <<< "$TRUSTED_REGISTRIES"
    for REGISTRY in "${REGISTRIES[@]}"; do
        CMD="$CMD --trusted-registry $REGISTRY"
    done
fi

# Set cloud keys and cloud tenant
# If CLOUD = true, then we use --report-sync and use these keys
export VET_API_KEY=$CLOUD_KEY
export VET_CONTROL_TOWER_TENANT_ID=$CLOUD_TENANT

# Run the Full command
$CMD

# Capture the exit status of the vet command immediately
VET_EXIT_CODE=$?

if [ ! -f "$ARTIFACT_FILENAME" ]; then
    echo "Artifact file not found: $ARTIFACT_FILENAME"
    exit 1
fi

# Exit with the status returned by vet
# If vet failed (non-zero), the pipe will now fail.
exit $VET_EXIT_CODE
