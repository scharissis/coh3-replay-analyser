#!/bin/bash
# Simple script to fetch the latest CoH3 data files from coh3-data repository

set -e

# Configuration
GITHUB_API="https://api.github.com/repos/cohstats/coh3-data/releases/latest"
BASE_URL="https://data.coh3stats.com/cohstats/coh3-data"
OUTPUT_DIR="data/coh3-data"

# Get latest tag
echo "Fetching latest release tag..."
LATEST_TAG=$(curl -s "$GITHUB_API" | grep '"tag_name":' | cut -d'"' -f4)
echo "Latest tag: $LATEST_TAG"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# List of data files to download
FILES=(
    "locstring.json"
    "sbps.json"
    "ebps.json" 
    "abilities.json"
    "battlegroup.json"
    "upgrade.json"
)

# List of locale files to download
LOCALE_FILES=(
    "locales/en-locstring.json"
)

# Download each file
echo "Downloading data files to $OUTPUT_DIR..."
for file in "${FILES[@]}"; do
    url="$BASE_URL/$LATEST_TAG/data/$file"
    echo "  Downloading $file..."
    curl -s -o "$OUTPUT_DIR/$file" "$url" || echo "  Warning: Failed to download $file"
done

# Download locale files
echo "Downloading locale files..."
for file in "${LOCALE_FILES[@]}"; do
    url="$BASE_URL/$LATEST_TAG/data/$file"
    echo "  Downloading $file..."
    # Create subdirectory if needed
    mkdir -p "$OUTPUT_DIR/$(dirname "$file")"
    curl -s -o "$OUTPUT_DIR/$file" "$url" || echo "  Warning: Failed to download $file"
done

# Create a metadata file with the tag info
echo "Creating metadata..."
ALL_FILES=("${FILES[@]}" "${LOCALE_FILES[@]}")
cat > "$OUTPUT_DIR/metadata.json" << EOF
{
  "tag": "$LATEST_TAG",
  "downloaded_at": "$(date -u +"%Y-%m-%dT%H:%M:%SZ")",
  "files": [
$(printf '    "%s"' "${ALL_FILES[0]}")
$(printf ',\n    "%s"' "${ALL_FILES[@]:1}")
  ]
}
EOF

echo "Download complete!"
echo "Files saved to: $OUTPUT_DIR"
echo "Data tag: $LATEST_TAG"

# Show file sizes
echo ""
echo "Downloaded files:"
for file in "${ALL_FILES[@]}"; do
    if [ -f "$OUTPUT_DIR/$file" ]; then
        size=$(ls -lh "$OUTPUT_DIR/$file" | awk '{print $5}')
        echo "  $file: $size"
    else
        echo "  $file: MISSING"
    fi
done