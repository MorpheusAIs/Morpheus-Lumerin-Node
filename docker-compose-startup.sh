#!/bin/sh

# Start the 'serve' command in the background
ollama serve &

# Optionally, wait for 'serve' to be fully ready. Adjust the sleep as necessary.
sleep 10 # This is a simple way. Consider more reliable checking mechanisms.

# Proceed with other commands
ollama pull llama2
ollama run llama2

# Keep the container running by tailing a file (or another long-running command)
tail -f /dev/null