#!/bin/bash

# Parse command line arguments
REMARKABLE_PATH="/home/root/.local/share/remarkable/xochitl"
SHOW_ICONS=""
SHOW_LABELS=""
SHOW_UUID=""
USE_COLOR="1"

while [ $# -gt 0 ]; do
    case "$1" in
        -icons)
            SHOW_ICONS="1"
            ;;
        -labels)
            SHOW_LABELS="1"
            ;;
        -uuid)
            SHOW_UUID="1"
            ;;
        -no-color)
            USE_COLOR=""
            ;;
        -*)
            echo "Unknown option: $1" >&2
            echo "Usage: $0 [path] [-icons] [-labels] [-uuid] [-no-color]" >&2
            exit 1
            ;;
        *)
            REMARKABLE_PATH="$1"
            ;;
    esac
    shift
done

if [ ! -d "$REMARKABLE_PATH" ]; then
    echo "Error: Path '$REMARKABLE_PATH' does not exist" >&2
    exit 1
fi

# Define colors
if [ -n "$USE_COLOR" ]; then
    COLOR_FOLDER="\033[36m"   # Cyan
    COLOR_PDF="\033[31m"      # Red
    COLOR_EPUB="\033[32m"     # Green
    COLOR_RESET="\033[0m"     # Reset
else
    COLOR_FOLDER=""
    COLOR_PDF=""
    COLOR_EPUB=""
    COLOR_RESET=""
fi

declare -A items
declare -A parents
declare -A types
declare -A names
declare -A doctypes

for metadata_file in "$REMARKABLE_PATH"/*.metadata; do
    [ -e "$metadata_file" ] || continue

    uuid=$(basename "$metadata_file" .metadata)

    # Skip deleted items
    if grep -q '"deleted"[[:space:]]*:[[:space:]]*true' "$metadata_file"; then
        continue
    fi

    name=$(grep '"visibleName"' "$metadata_file" | sed 's/.*"visibleName"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    type=$(grep '"type"' "$metadata_file" | sed 's/.*"type"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')
    parent=$(grep '"parent"' "$metadata_file" | sed 's/.*"parent"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/')

    [ -z "$name" ] && name="Unnamed"
    [ -z "$type" ] && type="DocumentType"

    items["$uuid"]=1
    names["$uuid"]="$name"
    types["$uuid"]="$type"
    parents["$uuid"]="$parent"

    # Detect document type (check EPUB first)
    if [ "$type" != "CollectionType" ]; then
        if [ -f "$REMARKABLE_PATH/$uuid.epub" ]; then
            doctypes["$uuid"]="epub"
        elif [ -f "$REMARKABLE_PATH/$uuid.pdf" ]; then
            doctypes["$uuid"]="pdf"
        else
            doctypes["$uuid"]="notebook"
        fi
    fi
done

# Function to print an item and its children
print_item() {
    local uuid="$1"
    local prefix="$2"
    local is_last="$3"
    local depth="$4"

    [ "$depth" -gt 50 ] && return
    [ -z "$uuid" ] && return
    [ -z "${items[$uuid]}" ] && return

    # Determine connector (we'll accept ├── for all items for now)
    local connector="├── "
    [ "$is_last" = "1" ] && connector="└── "

    # Determine icon, color, type label, and UUID display
    local icon="" color="" type_label="" uuid_display=""

    if [ "${types[$uuid]}" = "CollectionType" ]; then
        [ -n "$SHOW_ICONS" ] && icon="📁 "
        color="$COLOR_FOLDER"
        # No type label or UUID for folders
    else
        case "${doctypes[$uuid]}" in
            pdf)
                [ -n "$SHOW_ICONS" ] && icon="📕 "
                color="$COLOR_PDF"
                [ -n "$SHOW_LABELS" ] && type_label=" (pdf)"
                ;;
            epub)
                [ -n "$SHOW_ICONS" ] && icon="📗 "
                color="$COLOR_EPUB"
                [ -n "$SHOW_LABELS" ] && type_label=" (epub)"
                ;;
            *)
                [ -n "$SHOW_ICONS" ] && icon="📓 "
                # No color for notebooks (default terminal color)
                [ -n "$SHOW_LABELS" ] && type_label=" (notebook)"
                ;;
        esac
        # Add UUID for documents (not folders)
        [ -n "$SHOW_UUID" ] && uuid_display=" [$uuid]"
    fi

    echo -e "${prefix}${connector}${color}${icon}${names[$uuid]}${COLOR_RESET}${type_label}${uuid_display}"

    # Find and sort children
    local child_list=""
    for child_uuid in "${!items[@]}"; do
        if [ "${parents[$child_uuid]}" = "$uuid" ]; then
            # Create sort key: 0 for folders, 1 for documents, then name
            local sort_key="1"
            [ "${types[$child_uuid]}" = "CollectionType" ] && sort_key="0"
            child_list="${child_list}${sort_key}|${names[$child_uuid]}|${child_uuid}\n"
        fi
    done

    if [ -n "$child_list" ]; then
        # Sort children and convert to positional parameters
        local sorted_children=$(echo -e "${child_list}" | sort | cut -d'|' -f3)

        # Save current positional parameters
        local saved_params="$*"
        local saved_count="$#"

        # Convert sorted children to positional parameters
        set -- $sorted_children
        local total_children=$#
        local current_child=0

        # Process each child
        for child_uuid in "$@"; do
            [ -z "$child_uuid" ] && continue
            ((current_child++))

            local child_is_last="0"
            [ $current_child -eq $total_children ] && child_is_last="1"

            local new_prefix="${prefix}"
            if [ "$is_last" = "1" ]; then
                new_prefix="${prefix}    "
            else
                new_prefix="${prefix}│   "
            fi

            print_item "$child_uuid" "$new_prefix" "$child_is_last" $((depth + 1))
        done

        # Restore original positional parameters
        set -- $saved_params
    fi
}

# echo "reMarkable Filesystem"
# echo "===================="
echo "."

# Find root items and create sort list
root_list=""
for uuid in "${!items[@]}"; do
    parent="${parents[$uuid]}"
    if [ -z "$parent" ] || [ "$parent" = "trash" ] || [ -z "${items[$parent]}" ]; then
        # Create sort key
        sort_key="1"
        [ "${types[$uuid]}" = "CollectionType" ] && sort_key="0"
        root_list="${root_list}${sort_key}|${names[$uuid]}|${uuid}\n"
    fi
done

# Sort and print root items
if [ -n "$root_list" ]; then
    sorted_roots=$(echo -e "${root_list}" | sort | cut -d'|' -f3)

    # Convert to positional parameters
    set -- $sorted_roots
    total_roots=$#
    current_root=0

    # Process each root
    for uuid in "$@"; do
        [ -z "$uuid" ] && continue
        ((current_root++))

        is_last="0"
        [ $current_root -eq $total_roots ] && is_last="1"

        print_item "$uuid" "" "$is_last" 0
    done
fi
