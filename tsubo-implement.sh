#!/bin/bash
# Tsubo Implementation Script
# This script analyzes Tsubo (application) contracts and
# creates a plan for parallel AI implementation

set -e

# Color output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Usage
usage() {
    echo "Usage: $0 <tsubo.yaml>"
    echo ""
    echo "Example:"
    echo "  $0 ./poc/contracts/tsubo-todo-app.tsubo.yaml"
    exit 1
}

# Argument check
if [ $# -ne 1 ]; then
    usage
fi

TSUBO_FILE="$1"

if [ ! -f "$TSUBO_FILE" ]; then
    echo "Error: File not found: $TSUBO_FILE"
    exit 1
fi

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Tsubo Implementation Orchestrator${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Get Tsubo name
TSUBO_NAME=$(grep "name:" "$TSUBO_FILE" | head -1 | sed 's/.*name: *//' | tr -d '"')
echo -e "${GREEN}Tsubo: ${TSUBO_NAME}${NC}"
echo ""

# Get contracts directory path
CONTRACTS_DIR=$(dirname "$TSUBO_FILE")
echo -e "${BLUE}Contracts directory: ${CONTRACTS_DIR}${NC}"
echo ""

# ========================================
# Step 0: Verify context files
# ========================================
echo -e "${YELLOW}[Step 0] Verifying context files${NC}"

PROJECT_ROOT=$(cd "$CONTRACTS_DIR/../.." && pwd)
CONTEXT_FILES=(
    "$PROJECT_ROOT/PHILOSOPHY.md"
    "$PROJECT_ROOT/docs/DEVELOPMENT_PRINCIPLES.md"
    "$PROJECT_ROOT/docs/WHY_GO.md"
    "$PROJECT_ROOT/docs/CONTRACT_DESIGN.md"
)

MISSING_CONTEXT=0
for file in "${CONTEXT_FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "  ✓ $(basename $file)"
    else
        echo -e "  ✗ $(basename $file) not found"
        MISSING_CONTEXT=1
    fi
done

if [ $MISSING_CONTEXT -eq 1 ]; then
    echo -e "${YELLOW}Warning: Some context files are missing${NC}"
fi
echo ""

# ========================================
# Step 1: Enumerate objects (domains)
# ========================================
echo -e "${YELLOW}[Step 1] Enumerating objects${NC}"

# Extract objects from the objects section
OBJECTS=()
IN_OBJECTS=0

while IFS= read -r line; do
    if [[ "$line" =~ ^objects: ]]; then
        IN_OBJECTS=1
        continue
    fi

    if [ $IN_OBJECTS -eq 1 ]; then
        # End of objects section when indentation returns
        if [[ "$line" =~ ^[a-z] ]] && [[ ! "$line" =~ ^[[:space:]] ]]; then
            break
        fi

        # Find contract: lines
        if [[ "$line" =~ contract:.*\.object\.yaml ]]; then
            contract_file=$(echo "$line" | sed 's/.*contract: *//' | tr -d '"' | xargs)
            OBJECTS+=("$contract_file")
        fi
    fi
done < "$TSUBO_FILE"

if [ ${#OBJECTS[@]} -eq 0 ]; then
    echo "Error: No objects found"
    exit 1
fi

echo "Found ${#OBJECTS[@]} object(s):"
for obj in "${OBJECTS[@]}"; do
    echo "  - $obj"
done
echo ""

# ========================================
# Step 2: Analyze dependencies
# ========================================
echo -e "${YELLOW}[Step 2] Analyzing dependencies${NC}"

# Manage dependencies with temporary file
DEPS_FILE=$(mktemp)
trap "rm -f $DEPS_FILE" EXIT

for obj_file in "${OBJECTS[@]}"; do
    full_path="$CONTRACTS_DIR/$obj_file"
    obj_name=$(grep "name:" "$full_path" | head -1 | sed 's/.*name: *//' | tr -d '"')

    # Extract dependent services from dependencies.services section
    deps=""
    in_deps=0
    in_services=0
    while IFS= read -r line; do
        if [[ "$line" =~ ^dependencies: ]]; then
            in_deps=1
            continue
        fi

        if [ $in_deps -eq 1 ]; then
            # End of dependencies section
            if [[ "$line" =~ ^[a-z] ]] && [[ ! "$line" =~ ^[[:space:]] ]]; then
                break
            fi

            # Start of services section
            if [[ "$line" =~ ^[[:space:]]+services: ]]; then
                in_services=1
                continue
            fi

            # Start of other sections like databases (end of services)
            if [[ "$line" =~ ^[[:space:]]+[a-z]+: ]] && [[ ! "$line" =~ services: ]]; then
                in_services=0
                continue
            fi

            # Extract name within services section
            if [ $in_services -eq 1 ] && [[ "$line" =~ -[[:space:]]*name: ]]; then
                dep_name=$(echo "$line" | sed 's/.*name: *//' | tr -d '"' | xargs)
                if [ -n "$deps" ]; then
                    deps="$deps,$dep_name"
                else
                    deps="$dep_name"
                fi
            fi
        fi
    done < "$full_path"

    # Save to file (format: obj_name|deps)
    echo "$obj_name|$deps" >> "$DEPS_FILE"

    if [ -z "$deps" ]; then
        echo "  $obj_name: no dependencies"
    else
        echo "  $obj_name: depends on $deps"
    fi
done
echo ""

# ========================================
# Step 3: Determine implementation order
# ========================================
echo -e "${YELLOW}[Step 3] Determining implementation order${NC}"

# Function to get dependencies
get_deps() {
    local name=$1
    grep "^$name|" "$DEPS_FILE" | cut -d'|' -f2
}

# Simple implementation order (no dependencies → with dependencies)
WAVE_0=()
WAVE_1=()

for obj_file in "${OBJECTS[@]}"; do
    full_path="$CONTRACTS_DIR/$obj_file"
    obj_name=$(grep "name:" "$full_path" | head -1 | sed 's/.*name: *//' | tr -d '"')
    deps=$(get_deps "$obj_name")

    if [ -z "$deps" ]; then
        WAVE_0+=("$obj_file")
    else
        WAVE_1+=("$obj_file")
    fi
done

echo "Wave 0 (parallel execution - no dependencies):"
for obj in "${WAVE_0[@]}"; do
    obj_name=$(grep "name:" "$CONTRACTS_DIR/$obj" | head -1 | sed 's/.*name: *//' | tr -d '"')
    echo "  - $obj_name"
done

echo ""
echo "Wave 1 (execute after Wave 0 completes):"
for obj in "${WAVE_1[@]}"; do
    obj_name=$(grep "name:" "$CONTRACTS_DIR/$obj" | head -1 | sed 's/.*name: *//' | tr -d '"')
    deps=$(get_deps "$obj_name")
    echo "  - $obj_name (depends on: $deps)"
done
echo ""

# ========================================
# Step 4: Generate implementation plan
# ========================================
echo -e "${YELLOW}[Step 4] Generating implementation plan${NC}"

PLAN_FILE="/tmp/tsubo-implementation-plan.json"

cat > "$PLAN_FILE" <<EOF
{
  "tsubo": "$TSUBO_NAME",
  "tsubo_file": "$TSUBO_FILE",
  "contracts_dir": "$CONTRACTS_DIR",
  "project_root": "$PROJECT_ROOT",
  "context_files": [
EOF

# Add context files to JSON array
first=true
for file in "${CONTEXT_FILES[@]}"; do
    if [ -f "$file" ]; then
        if [ "$first" = true ]; then
            echo "    \"$file\"" >> "$PLAN_FILE"
            first=false
        else
            echo "    ,\"$file\"" >> "$PLAN_FILE"
        fi
    fi
done

cat >> "$PLAN_FILE" <<EOF
  ],
  "waves": [
    {
      "wave": 0,
      "parallel": true,
      "objects": [
EOF

# Wave 0 objects
first=true
for obj in "${WAVE_0[@]}"; do
    full_path="$CONTRACTS_DIR/$obj"
    obj_name=$(grep "name:" "$full_path" | head -1 | sed 's/.*name: *//' | tr -d '"')

    if [ "$first" = true ]; then
        first=false
    else
        echo "        ," >> "$PLAN_FILE"
    fi

    cat >> "$PLAN_FILE" <<EOF
        {
          "name": "$obj_name",
          "contract": "$full_path",
          "dependencies": []
        }
EOF
done

cat >> "$PLAN_FILE" <<EOF

      ]
    },
    {
      "wave": 1,
      "parallel": true,
      "objects": [
EOF

# Wave 1 objects
first=true
for obj in "${WAVE_1[@]}"; do
    full_path="$CONTRACTS_DIR/$obj"
    obj_name=$(grep "name:" "$full_path" | head -1 | sed 's/.*name: *//' | tr -d '"')
    deps=$(get_deps "$obj_name")

    if [ "$first" = true ]; then
        first=false
    else
        echo "        ," >> "$PLAN_FILE"
    fi

    # Convert dependencies to array
    dep_array="["
    if [ -n "$deps" ]; then
        IFS=',' read -ra DEP_ARRAY <<< "$deps"
        first_dep=true
        for dep in "${DEP_ARRAY[@]}"; do
            if [ "$first_dep" = true ]; then
                dep_array="${dep_array}\"$dep\""
                first_dep=false
            else
                dep_array="${dep_array},\"$dep\""
            fi
        done
    fi
    dep_array="${dep_array}]"

    cat >> "$PLAN_FILE" <<EOF
        {
          "name": "$obj_name",
          "contract": "$full_path",
          "dependencies": $dep_array
        }
EOF
done

cat >> "$PLAN_FILE" <<EOF

      ]
    }
  ]
}
EOF

echo -e "${GREEN}Implementation plan generated: $PLAN_FILE${NC}"
echo ""

# ========================================
# Step 5: Summary
# ========================================
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Ready for Implementation${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Tsubo: $TSUBO_NAME"
echo "Number of objects: ${#OBJECTS[@]}"
echo "Implementation plan: $PLAN_FILE"
echo ""
echo -e "${GREEN}Next steps:${NC}"
echo "1. Review the plan: cat $PLAN_FILE | jq"
echo "2. Start parallel implementation with AI agents"
echo ""
echo -e "${YELLOW}Each AI agent will receive:${NC}"
echo "  - Tsubo philosophy (PHILOSOPHY.md)"
echo "  - Development principles (DEVELOPMENT_PRINCIPLES.md)"
echo "  - Why Go language (WHY_GO.md)"
echo "  - Contract design (CONTRACT_DESIGN.md)"
echo "  - Object contract (.object.yaml)"
echo ""
