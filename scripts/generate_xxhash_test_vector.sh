#!/bin/bash

# Script to generate a xxHash32 test vector for a single input
# Usage: ./generate_xxhash_test_vector.sh "input_string"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <input_string>"
  echo "Example: $0 \"abc\""
  echo "Example: $0 \"hello world\""
  exit 1
fi

INPUT="$1"
OUTPUT_DIR="mocks/testdata/xxhash32"

HASH=$(echo -n $INPUT | xxhsum -H0 - | cut -d ' ' -f 1)
OUTPUT_FILE="${OUTPUT_DIR}/xxhash32_${HASH}.json"

# Create JSON file
cat > "$OUTPUT_FILE" << EOF
{
  "input": "$INPUT",
  "seed": "0",
  "expected_hash": "$HASH"
}
EOF

echo "Test vector for input='$INPUT', seed=$SEED written to $OUTPUT_FILE"
