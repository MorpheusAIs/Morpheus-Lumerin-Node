#!/bin/bash

BIN_PATH="./node_modules/.bin"

# Check if the executables exist
if [ ! -x "$BIN_PATH/hardhat" ]; then
  echo "Hardhat is not found or not executable in $BIN_PATH."
  exit 1
fi

# Function to stop the hardhat process
stop_hardhat() {
  echo "Stopping hardhat process due to error."
  kill $HARDHAT_PID
  exit 1
}

# Start the hardhat process in the background
$BIN_PATH/hardhat node &
HARDHAT_PID=$!
echo "Started hardhat with PID $HARDHAT_PID"

# Run the migration process
$BIN_PATH/hardhat run --network localhost scripts/deploy-local.ts
MIGRATION_EXIT_CODE=$?

# Check if migration was successful
if [ $MIGRATION_EXIT_CODE -eq 0 ]; then
  echo "Migration successful. hardhat process is still running in the background."
else
  echo "Migration failed with exit code $MIGRATION_EXIT_CODE"
  stop_hardhat
fi

# Wait for the hardhat process to exit
wait $HARDHAT_PID
HARDHAT_EXIT_CODE=$?

# Check the exit status of the hardhat process
if [ $HARDHAT_EXIT_CODE -eq 0 ]; then
  echo "Hardhat process exited successfully."
else
  echo "Hardhat process exited with exit code $HARDHAT_EXIT_CODE"
fi