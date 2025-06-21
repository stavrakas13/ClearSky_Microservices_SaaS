#!/usr/bin/env bash
# save as gather_frontend_and_backend_code.sh and run: bash gather_frontend_and_backend_code.sh

OUTPUT="all_code.txt"

# Remove old output if it exists
rm -f "$OUTPUT"

echo "Collecting frontend code..." >> "$OUTPUT"
echo "=========================" >> "$OUTPUT"
echo "" >> "$OUTPUT"

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

echo "" >> "$OUTPUT"
echo "Collecting backend Go code..." >> "$OUTPUT"
echo "=============================" >> "$OUTPUT"
echo "" >> "$OUTPUT"

# List of Go service directories
GO_DIRS=(
  "credits_service"
  "final_grades"
  "google_auth_service"
  "initial_grades"
  "instructor_review_reply_service"
  "orchestrator"
  "registration_service"
  "stats_service"
  "student_request_review_service"
  "user_management_service"
  "View_personal_grades"
)

# Process each Go directory
for dir in "${GO_DIRS[@]}"; do
  if [ -d "$dir" ]; then
    echo "Processing directory: $dir" >> "$OUTPUT"
    echo "----------------------------" >> "$OUTPUT"
    echo "" >> "$OUTPUT"
    
    # Find and process all .go files in the directory
    find "$dir" \
      -type f -iname "*.go" -print \
      | sort \
      | while IFS= read -r file; do
          echo "===== $file =====" >> "$OUTPUT"
          cat "$file" >> "$OUTPUT"
          echo "" >> "$OUTPUT"
        done
    
    echo "" >> "$OUTPUT"
  else
    echo "Warning: Directory '$dir' not found" >> "$OUTPUT"
    echo "" >> "$OUTPUT"
  fi
done

echo "All frontend and backend code collected into $OUTPUT."