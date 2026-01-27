#!/usr/bin/env bash

# ==============================================================================
# SafeDep Vet Pipe
# ==============================================================================

# Global Configuration
ARTIFACT_FILENAME="vet-report.json"
DEFAULT_POLICY="/default-policy.yml"
VET_CMD_ARGS=() # Array to hold command arguments
VET_EXIT_CODE=0

# ------------------------------------------------------------------------------
# 1. Initialization and Defaults
# ------------------------------------------------------------------------------
init_defaults() {
    echo ""
    echo "Executing SafeDep Vet Pipe..."
    echo ""

    # Set Default values using parameter expansion
    export VET_VERSION=${VET_VERSION:-"latest"}
    export CLOUD=${CLOUD:-"false"}
    export TIMEOUT=${TIMEOUT:-300}
    export SKIP_FILTER_CI_FAIL=${SKIP_FILTER_CI_FAIL:-"false"}
}

# ------------------------------------------------------------------------------
# 2. Validation Checks
# ------------------------------------------------------------------------------
check_prerequisites() {
    echo "[+] Checking prerequisites..."

    if ! command -v vet &> /dev/null; then
        echo "Error: 'vet' command is not installed or not in PATH."
        exit 1
    fi

    if [ -n "$POLICY" ] && [ ! -f "$POLICY" ]; then
        echo "Error: Policy file not found at: $POLICY"
        exit 1
    fi

    if [ -n "$EXCEPTION_FILE" ] && [ ! -f "$EXCEPTION_FILE" ]; then
        echo "Error: Exception file not found at: $EXCEPTION_FILE"
        exit 1
    fi
}

validate_cloud_config() {
    if [ "$CLOUD" = "true" ]; then
        echo "[+] Validating cloud configuration..."
        if [ -z "$CLOUD_KEY" ] || [ -z "$CLOUD_TENANT" ]; then
            echo "Error: CLOUD_KEY and CLOUD_TENANT must be provided when CLOUD is enabled."
            exit 1
        fi

        # Export Cloud specific Env Vars needed for execution
        export VET_API_KEY=$CLOUD_KEY
        export VET_CONTROL_TOWER_TENANT_ID=$CLOUD_TENANT
    fi
}

# ------------------------------------------------------------------------------
# 3. Command Construction Functions
# ------------------------------------------------------------------------------

# Adds the base scan arguments
add_base_config() {
    VET_CMD_ARGS+=( "scan" "-s" "--report-json" "$ARTIFACT_FILENAME" "--fail-fast" )
}

# Adds policy enforcement arguments
add_policy_config() {
    if [ -n "$POLICY" ]; then
        echo "[+] Applying Policy: $POLICY"
        VET_CMD_ARGS+=( "--filter-suite" "$POLICY" )
    else
        echo "[+] Applying Default Policy: $DEFAULT_POLICY"
        VET_CMD_ARGS+=( "--filter-suite" "$DEFAULT_POLICY" )
    fi

    if [ "$SKIP_FILTER_CI_FAIL" = "false" ]; then
        VET_CMD_ARGS+=( "--filter-fail" )
    fi
}

# Adds exception file arguments
add_exceptions_config() {
    if [ -n "$EXCEPTION_FILE" ]; then
        echo "[+] Applying Exceptions: $EXCEPTION_FILE"
        VET_CMD_ARGS+=( "--exceptions" "$EXCEPTION_FILE" )
    fi
}

# Adds cloud reporting and malware analysis arguments
add_cloud_features() {
    if [ "$CLOUD" = "true" ]; then
        echo "[+] Enabling Cloud Sync and Malware Analysis"
        VET_CMD_ARGS+=(
            "--report-sync"
            "--report-sync-project" "$BITBUCKET_REPO_FULL_NAME"
            "--report-sync-project-version" "$BITBUCKET_BRANCH"
            "--malware"
            "--malware-analysis-timeout" "${TIMEOUT}s"
        )
    fi
}

# Adds trusted registries from CSV list
add_registry_config() {
    if [ -n "$TRUSTED_REGISTRIES" ]; then
        echo "[+] Configuring Trusted Registries"
        IFS=',' read -ra REGISTRIES <<< "$TRUSTED_REGISTRIES"
        for REGISTRY in "${REGISTRIES[@]}"; do
            # Trim whitespace just in case
            REGISTRY=$(echo "$REGISTRY" | xargs)
            VET_CMD_ARGS+=( "--trusted-registry" "$REGISTRY" )
        done
    fi
}

# Generate Exception file for Pull Reqeust changed pacakges scanning
generate_pr_exceptions() {
    # Verify this is a PR! If we have this variable set by bitbucket, its a PR
    if [ -n "$BITBUCKET_PR_ID" ]; then
        echo "Running vet scan in Pull Request"

        # 0. fixes ownership issues when this pipe created file
        # fatal: detected dubious ownership in repository at '/opt/atlassian/pipelines/agent/build'
        git config --global --add safe.directory /opt/atlassian/pipelines/agent/build

        # 1. Fetch the Base Branch data
        git fetch origin $BITBUCKET_PR_DESTINATION_BRANCH

        # 2. Switch to the Base Branch context
        git checkout -f $BITBUCKET_PR_DESTINATION_BRANCH

        # 2.5 Set reusable variables
        export VET_JSON_DUMP_DIR="/tmp/safedep-vet/dump/"
        export VET_PR_EXCEPTION_FILE_PATH="/tmp/safedep-vet/exception.yml"

        # 3. Create your vet json dump folder
        mkdir -p $VET_JSON_DUMP_DIR
        # Silent console reporting on intermetidery vet commands
        vet scan --no-banner --report-summary=false --silent --json-dump-dir $VET_JSON_DUMP_DIR --enrich=false .

        # 4. Generate Exceptions
        # Silent console reporting on intermetidery vet commands
        vet query --no-banner --report-summary=false --from $VET_JSON_DUMP_DIR --exceptions-filter true --exceptions-generate $VET_PR_EXCEPTION_FILE_PATH

        # 5. Switch back to your Feature (Head) Branch
        git checkout -f $BITBUCKET_BRANCH

        # 6 Add Extra Exceptions File  `--exceptions-extra` for PR
        # `--exceptions` flag is already used for $EXCEPTION_FILE input variable
        # Hence we will use this extra exception flag for PR changes packages exception logic, both work together their files are logically concatinated.
        VET_CMD_ARGS+=( "--exceptions-extra" $VET_PR_EXCEPTION_FILE_PATH )
    else
        echo "Running vet scan in Push."
        # No Extra Logic is required, we do full scan
    fi
}

# ------------------------------------------------------------------------------
# 4. Execution and Verification
# ------------------------------------------------------------------------------
execute_scan() {
    echo "----------------------------------------------------------------"
    echo "Running command: vet ${VET_CMD_ARGS[*]}"
    echo "----------------------------------------------------------------"

    # Run the command using the array
    "vet" "${VET_CMD_ARGS[@]}"

    # Capture exit code
    VET_EXIT_CODE=$?
}

verify_artifact() {
    if [ ! -f "$ARTIFACT_FILENAME" ]; then
        echo "Error: Artifact file was not generated: $ARTIFACT_FILENAME"
        exit 1
    fi
}

# ------------------------------------------------------------------------------
# Main Orchestration
# ------------------------------------------------------------------------------
main() {
    init_defaults

    # Validation Phase
    check_prerequisites
    validate_cloud_config

    # Build the command arguments
    add_base_config
    add_policy_config
    add_exceptions_config
    add_cloud_features
    add_registry_config

    # Generate Exception file for Pull Request changed packages scanning
    generate_pr_exceptions

    # Execution Phase
    execute_scan

    # Post-Execution Phase
    verify_artifact

    # Call upload_report.sh file to create and update Bitbucket Code Insights Report
    /upload_report.sh $ARTIFACT_FILENAME

    # Exit with the code returned by the vet tool
    exit $VET_EXIT_CODE
}

# Invoke Main
main
