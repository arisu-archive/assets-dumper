#!/bin/bash

INPUT_FILE="$1"
if [ -z "$INPUT_FILE" ]; then
  INPUT_FILE="mocks/testdata/xxhash32/inputs.txt"
fi

while IFS= read -r line; do
  echo "Processing: $line"
  ./scripts/generate_xxhash_test_vector.sh "$line"
done < $INPUT_FILE
