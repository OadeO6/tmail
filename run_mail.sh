#!/bin/env bash

# Script to run mail server

# Configuration script that handles environment loading based on mode
# Usage: ./run_mail.sh [dev|prod]
# Defaults to dev mode if no argument is provided
# The script will run "go run mail/main.go" with the appropriate environment

# Setup logging function with timestamps and log levels


RUN_COMMAND="go run mail/main.go"

log() {
    local level=$1
    local message=$2
    local timestamp=$(date "+%Y-%m-%d %H:%M:%S")
    echo "[$timestamp] [$level] $message"
}

# Function to load environment variables from .env file
load_env_from_file() {
    local env_file=".env"

    # Check if .env file exists
    if [ ! -f "$env_file" ]; then
        log "ERROR" "Environment file '$env_file' not found!"
        return 1
    fi

    log "INFO" "Loading environment variables from $env_file"

    # Read each line from .env file
    while IFS= read -r line || [[ -n "$line" ]]; do
        # Skip comments and empty lines
        [[ $line =~ ^#.*$ ]] && continue
        [[ -z $line ]] && continue

        # Export the variable
        export "$line"

        # Get variable name for logging (without exposing the value)
        var_name=$(echo "$line" | cut -d '=' -f 1)
        log "DEBUG" "Loaded environment variable: $var_name"
    done < "$env_file"

    log "INFO" "Environment variables loaded successfully"
    return 0
}

# Main function to handle script execution
main() {
    # Check if mode argument is provided
    # Get the mode from first argument
    local mode=$1
    if [ $# -lt 1 ]; then
        log "INFO" "No argument provided defaulting to dev mode"
    	local mode=dev
    fi

    shift # Remove the first argument, leaving the command to run

    case "$mode" in
        "dev")
            log "INFO" "Running in DEVELOPMENT mode"
            if ! load_env_from_file; then
                log "ERROR" "Failed to load environment variables. Exiting."
                exit 1
            fi
            ;;
        "prod")
            log "INFO" "Running in PRODUCTION mode - using system environment"
            ;;
        *)
            log "ERROR" "Invalid mode: $mode. Use 'dev' or 'prod'"
            exit 1
            ;;
    esac

    # Execute the Go application
    log "INFO" "Executing: go run mail/main.go"
    $RUN_COMMAND
}

# Execute main function with all arguments
main "$@"
