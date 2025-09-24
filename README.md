# rmtree

A tree-style display tool for the reMarkable tablet's document filesystem, showing the hierarchical structure of documents and folders.

### Automatic Installation (Recommended)

> [!CAUTION]
> Piping code from the internet directly into `bash` can be dangerous. Make sure you trust the source and know what it will do to your system.

The easiest way to install is using the installation script that automatically detects your device architecture:

```bash
wget -qO- https://github.com/rmitchellscott/rm-tree/raw/main/install.sh | bash
```

## Manual Installation

Download the correct binary for your device:

```bash
# For reMarkable 1/2 (ARMv7):
wget https://github.com/rmitchellscott/rm-tree/releases/latest/download/rmtree-armv7.tar.gz
tar -xzf rmtree-armv7.tar.gz
mv rmtree-armv7 rmtree
chmod +x rmtree

# For reMarkable Paper Pro & Paper Pro Move (ARM64):
wget https://github.com/rmitchellscott/rm-tree/releases/latest/download/rmtree-aarch64.tar.gz
tar -xzf rmtree-aarch64.tar.gz
mv rmtree-aarch64 rmtree
chmod +x rmtree
```

## Usage

```bash
./rmtree [path] [options]
```

**Default path**: `/home/root/.local/share/remarkable/xochitl`

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
