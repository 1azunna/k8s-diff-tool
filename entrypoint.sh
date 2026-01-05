#!/bin/sh
set -e

# Initialize arguments
ARGS=""

# Handle flags based on inputs
if [ "$INPUT_DIRECTORY" = "true" ]; then
  ARGS="$ARGS -d"
fi

if [ "$INPUT_SECURE_MODE" = "true" ]; then
  ARGS="$ARGS -s"
fi

if [ -n "$INPUT_INCLUDE" ]; then
  ARGS="$ARGS -i $INPUT_INCLUDE"
fi

if [ -n "$INPUT_EXCLUDE" ]; then
  ARGS="$ARGS -e $INPUT_EXCLUDE"
fi

# Add positional arguments (Paths)
if [ -z "$INPUT_BASE_PATH" ] || [ -z "$INPUT_HEAD_PATH" ]; then
  echo "Error: base_path and head_path are required."
  exit 1
fi

ARGS="$ARGS $INPUT_BASE_PATH $INPUT_HEAD_PATH"

echo "Running: kdiff $ARGS"
# Execute the tool
# Execute the tool and capture output
# We assume the binary is located at $GITHUB_ACTION_PATH/bin/kdiff
OUTPUT=$($GITHUB_ACTION_PATH/bin/kdiff $ARGS)
EXIT_CODE=$?

# Print output to console for logging
echo "$OUTPUT"

# Write output to GITHUB_OUTPUT with multiline support
EOF=$(dd if=/dev/urandom bs=15 count=1 status=none | base64)
echo "diff<<$EOF" >> $GITHUB_OUTPUT
echo "$OUTPUT" >> $GITHUB_OUTPUT
echo "$EOF" >> $GITHUB_OUTPUT

exit $EXIT_CODE
