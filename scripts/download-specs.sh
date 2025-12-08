#!/bin/bash
set -euo pipefail

# GoFHIR - FHIR Specification Downloader
# Downloads StructureDefinitions, ValueSets, and CodeSystems from hl7.org/fhir

SPECS_DIR="./specs"
BASE_URL="https://hl7.org/fhir"

# FHIR versions and their URLs
declare -A FHIR_VERSIONS=(
    ["r4"]="R4"
    ["r4b"]="R4B"
    ["r5"]="R5"
)

declare -A FHIR_URLS=(
    ["r4"]="https://hl7.org/fhir/R4"
    ["r4b"]="https://hl7.org/fhir/R4B"
    ["r5"]="https://hl7.org/fhir/R5"
)

# Files to download for each version
FILES=(
    "definitions.json.zip"
    "valuesets.json.zip"
)

download_version() {
    local version=$1
    local url=${FHIR_URLS[$version]}
    local output_dir="$SPECS_DIR/$version"

    echo "Downloading FHIR ${FHIR_VERSIONS[$version]} specifications..."
    mkdir -p "$output_dir"

    for file in "${FILES[@]}"; do
        local file_url="$url/$file"
        local output_file="$output_dir/$file"

        echo "  Downloading $file..."
        if curl -fsSL "$file_url" -o "$output_file"; then
            echo "  Extracting $file..."
            unzip -q -o "$output_file" -d "$output_dir"
            rm "$output_file"
        else
            echo "  Warning: Failed to download $file"
        fi
    done

    # Download individual important files
    echo "  Downloading profiles-resources.json..."
    curl -fsSL "$url/profiles-resources.json" -o "$output_dir/profiles-resources.json" || true

    echo "  Downloading profiles-types.json..."
    curl -fsSL "$url/profiles-types.json" -o "$output_dir/profiles-types.json" || true

    echo "  Done with FHIR ${FHIR_VERSIONS[$version]}"
    echo ""
}

cleanup() {
    echo "Cleaning up temporary files..."
    find "$SPECS_DIR" -name "*.zip" -delete 2>/dev/null || true
}

main() {
    local version="${1:-all}"

    echo "GoFHIR - FHIR Specification Downloader"
    echo "======================================"
    echo ""

    if [ "$version" = "all" ]; then
        for v in "${!FHIR_VERSIONS[@]}"; do
            download_version "$v"
        done
    elif [[ -v FHIR_VERSIONS[$version] ]]; then
        download_version "$version"
    else
        echo "Error: Unknown version '$version'"
        echo "Available versions: ${!FHIR_VERSIONS[*]}"
        exit 1
    fi

    cleanup

    echo "Download complete!"
    echo ""
    echo "Specifications are stored in:"
    ls -la "$SPECS_DIR"/*/
}

main "$@"
