#!/bin/bash

# Script to add Jekyll front matter and navigation to all documentation files

DOCS_DIR="/Users/prasenjit/GolandProjects/openid-golang/docs"

# Navigation header
NAV_HEADER="[🏠 Home](index.md) | [📚 All Docs](index.md#-quick-navigation) | [⚡ Quick Start](QUICKSTART.md) | [🐳 Docker](DOCKER.md) | [📖 API](API.md)

---

"

# Function to add front matter if not present
add_front_matter() {
    local file="$1"
    local title="$2"

    # Check if file already has front matter
    if head -n 1 "$file" | grep -q "^---$"; then
        echo "Skipping $file (already has front matter)"
        return
    fi

    # Create temp file
    local temp_file=$(mktemp)

    # Get first heading as title if not provided
    if [ -z "$title" ]; then
        title=$(grep -m 1 "^#" "$file" | sed 's/^#* *//')
    fi

    # Add front matter
    {
        echo "---"
        echo "layout: default"
        echo "title: $title"
        echo "---"
        echo ""
        echo "$NAV_HEADER"
        cat "$file"
    } > "$temp_file"

    # Replace original file
    mv "$temp_file" "$file"
    echo "Added front matter to $file"
}

# Process all markdown files except index.md and INDEX.md
find "$DOCS_DIR" -maxdepth 1 -name "*.md" -type f | while read file; do
    filename=$(basename "$file")

    # Skip index files and files that already have front matter
    if [ "$filename" = "index.md" ] || [ "$filename" = "INDEX.md" ]; then
        continue
    fi

    add_front_matter "$file"
done

echo "Done!"

