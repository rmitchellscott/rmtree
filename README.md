# rmtree

A tree-style display tool for the reMarkable tablet's document filesystem, showing the hierarchical structure of documents and folders.

## Installation

On your reMarkable tablet:

```bash
# Download the script
wget https://raw.githubusercontent.com/rmitchellscott/rmtree/refs/heads/main/rmtree.sh

# Make it executable
chmod +x rmtree.sh
```

## Usage

```bash
./rmtree.sh [path] [options]
``` 
**Default path**: `/home/root/.local/share/remarkable/xochitl`

The script will take several seconds to run.

## Options

- `-icons` - Show emoji icons (📁 📕 📗 📓)
- `-labels` - Show document type labels (pdf), (epub), (notebook)
- `-uuid` - Show document UUIDs in square brackets (documents only, not folders)
- `-no-color` - Disable colored output

## Examples

**Default** (clean, colored):
```
.
├── Books
│   └──Project Hail Mary
├── Calendar
│   └── Calendar-2025
└── To Do
```

**With labels** (`-labels`):
```
.
├── Books
│   └── Project Hail Mary (epub)
├── Calendar
│   └── Calendar-2025 (pdf)
└── To Do (notebook)
```

**With icons and labels** (`-icons -labels`):
```
.
├── 📁 Books
│   └── 📗 Project Hail Mary (epub)
├── 📁 Calendar
│   └── 📕 Calendar-2025 (pdf)
└── 📓 To Do (notebook)
```

**With UUIDs** (`-uuid`):
```
.
├── Books
│   └──Project Hail Mary [3f05b2d1-90e0-458a-b233-7966564d2172]
├── Calendar
│   └── Calendar-2025 [67f60935-7978-4fe4-b234-64b70ed17c3e]
└── To Do [d1a44483-3023-4b16-b677-ea75211252ca]
```
