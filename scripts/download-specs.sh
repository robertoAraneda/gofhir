#!/bin/bash
set -eo pipefail

# GoFHIR - FHIR Specification Downloader
# Downloads StructureDefinitions from hl7.org/fhir

SPECS_DIR="./specs"

download_r4() {
    local output_dir="$SPECS_DIR/r4"
    local url="https://hl7.org/fhir/R4"

    echo "Downloading FHIR R4 specifications..."
    mkdir -p "$output_dir"

    echo "  Downloading definitions.json.zip..."
    if curl -fsSL "$url/definitions.json.zip" -o "$output_dir/definitions.json.zip"; then
        echo "  Extracting..."
        unzip -q -o "$output_dir/definitions.json.zip" -d "$output_dir"
        rm "$output_dir/definitions.json.zip"
    fi

    echo "  Downloading profiles-types.json..."
    curl -fsSL "$url/profiles-types.json" -o "$output_dir/profiles-types.json" || true

    echo "  Downloading profiles-resources.json..."
    curl -fsSL "$url/profiles-resources.json" -o "$output_dir/profiles-resources.json" || true

    echo "  Done with FHIR R4"
}

download_r4b() {
    local output_dir="$SPECS_DIR/r4b"
    local url="https://hl7.org/fhir/R4B"

    echo "Downloading FHIR R4B specifications..."
    mkdir -p "$output_dir"

    echo "  Downloading definitions.json.zip..."
    if curl -fsSL "$url/definitions.json.zip" -o "$output_dir/definitions.json.zip"; then
        echo "  Extracting..."
        unzip -q -o "$output_dir/definitions.json.zip" -d "$output_dir"
        rm "$output_dir/definitions.json.zip"
    fi

    echo "  Downloading profiles-types.json..."
    curl -fsSL "$url/profiles-types.json" -o "$output_dir/profiles-types.json" || true

    echo "  Downloading profiles-resources.json..."
    curl -fsSL "$url/profiles-resources.json" -o "$output_dir/profiles-resources.json" || true

    echo "  Done with FHIR R4B"
}

download_r5() {
    local output_dir="$SPECS_DIR/r5"
    local url="https://hl7.org/fhir/R5"

    echo "Downloading FHIR R5 specifications..."
    mkdir -p "$output_dir"

    echo "  Downloading definitions.json.zip..."
    if curl -fsSL "$url/definitions.json.zip" -o "$output_dir/definitions.json.zip"; then
        echo "  Extracting..."
        unzip -q -o "$output_dir/definitions.json.zip" -d "$output_dir"
        rm "$output_dir/definitions.json.zip"
    fi

    echo "  Downloading profiles-types.json..."
    curl -fsSL "$url/profiles-types.json" -o "$output_dir/profiles-types.json" || true

    echo "  Downloading profiles-resources.json..."
    curl -fsSL "$url/profiles-resources.json" -o "$output_dir/profiles-resources.json" || true

    echo "  Done with FHIR R5"
}

main() {
    local version="${1:-all}"

    echo "GoFHIR - FHIR Specification Downloader"
    echo "======================================"
    echo ""

    case "$version" in
        r4)
            download_r4
            ;;
        r4b)
            download_r4b
            ;;
        r5)
            download_r5
            ;;
        all)
            download_r4
            download_r4b
            download_r5
            ;;
        *)
            echo "Error: Unknown version '$version'"
            echo "Available versions: r4, r4b, r5, all"
            exit 1
            ;;
    esac

    echo ""
    echo "Download complete!"
    echo ""
    echo "Specifications stored in: $SPECS_DIR"
    ls -la "$SPECS_DIR"/
}

main "$@"
