#!/usr/bin/env bash
# save as gather_frontend_code.sh and run: bash gather_frontend_code.sh

OUTPUT="all_frontend_code.txt"

# Remove old output if it exists
rm -f "$OUTPUT"

# Find and process each .js, .ejs, .css file, excluding node_modules
find front-end \
  -path "*/node_modules" -prune -o \
  -type f \( -iname "*.js" -o -iname "*.ejs" -o -iname "*.css" \) -print \
  | sort \
  | while IFS= read -r file; do
      echo "===== $file =====" >> "$OUTPUT"
      cat "$file" >> "$OUTPUT"
      echo "" >> "$OUTPUT"
    done

echo "All frontend code collected into $OUTPUT."
