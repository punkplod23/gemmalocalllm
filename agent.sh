#!/bin/bash

# ==============================================================================
# Agentic Disk Monitoring and Docker Pruning Script
#
# This script acts as a simple agent to autonomously monitor disk space and
# perform cleanup actions when a problem is detected. It includes observation,
# decision, action, and reporting steps.
# ==============================================================================

# --- Configuration ---
# Set the disk usage threshold (in percent) that triggers the cleanup action.
# For example, a value of 80 means the agent will act if disk usage is >= 80%.
THRESHOLD=80

# The filesystem to monitor.
FILESYSTEM="/"

# --- Function Definitions ---

# Step 1: OBSERVE
# Gathers the current disk usage percentage for the specified filesystem.
function observe_disk_usage() {
  local usage=$(df -f "${FILESYSTEM}" | grep '/' | awk '{print $5}' | tr -d '%')
  echo $usage
}

# Step 2: ACT
# Executes the docker-compose system prune command to free up space.
function perform_cleanup() {
  echo "--- ACTION: Initiating cleanup process ---"
  # The `-f` flag is critical for non-interactive environments, as it
  # forces the prune operation without a confirmation prompt.
  # This command removes all unused containers, networks, images, and build cache.
  docker-compose system prune -f
  echo "--- ACTION: Cleanup complete ---"
}

# Step 3: REPORT
# Compares the before and after disk usage and reports the outcome.
function report_status() {
  local before_usage=$1
  local after_usage=$(observe_disk_usage)

  echo "--- REPORT: Agent action summary ---"
  echo "Disk usage before cleanup: ${before_usage}%"
  echo "Disk usage after cleanup: ${after_usage}%"
  local freed_space=$((before_usage - after_usage))
  echo "Space freed: ${freed_space}%"
  echo "-------------------------------------"
}

# --- Main Agentic Logic ---

echo "--- MCP Agent starting autonomous cycle ---"

# Step 1: OBSERVE the current state.
initial_usage=$(observe_disk_usage)
echo "OBSERVATION: Current disk usage on ${FILESYSTEM} is ${initial_usage}%"

# Step 2: DECIDE based on observation.
if [[ ${initial_usage} -ge ${THRESHOLD} ]]; then
  echo "DECISION: Disk usage exceeds threshold (${THRESHOLD}%). Action required."
  
  # Step 3: ACT on the decision.
  perform_cleanup
  
  # Step 4: REPORT on the outcome of the action.
  report_status "${initial_usage}"
else
  echo "DECISION: Disk usage is below threshold (${THRESHOLD}%). No action needed."
fi

echo "--- MCP Agent cycle finished ---"