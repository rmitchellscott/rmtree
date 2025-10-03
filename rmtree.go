package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var version = "dev"

type Metadata struct {
	VisibleName string `json:"visibleName"`
	Type        string `json:"type"`
	Parent      string `json:"parent"`
	Deleted     bool   `json:"deleted"`
}

type Item struct {
	UUID     string
	Name     string
	Type     string
	Parent   string
	DocType  string
	SortKey  string
}

type Config struct {
	Path      string
	ShowIcons bool
	ShowLabels bool
	ShowUUID  bool
	UseColor  bool
}

var colors = map[string]string{
	"folder": "\033[36m",
	"pdf":    "\033[31m",
	"epub":   "\033[32m",
	"reset":  "\033[0m",
}

func main() {
	config := parseArgs()

	if _, err := os.Stat(config.Path); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: Path '%s' does not exist\n", config.Path)
		os.Exit(1)
	}

	items, err := loadItems(config.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading items: %v\n", err)
		os.Exit(1)
	}

	children := buildChildrenMap(items)
	sortItems(items, children)

	printTree(items, children, config)
}

func parseArgs() Config {
	config := Config{
		Path:     "/home/root/.local/share/remarkable/xochitl",
		UseColor: true,
	}

	flag.BoolVar(&config.ShowIcons, "icons", false, "Show emoji icons")
	flag.BoolVar(&config.ShowLabels, "labels", false, "Show document type labels")
	flag.BoolVar(&config.ShowUUID, "uuid", false, "Show document UUIDs")
	noColor := flag.Bool("no-color", false, "Disable colored output")
	showVersion := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *showVersion {
		fmt.Println("rmtree version", version)
		os.Exit(0)
	}

	if flag.NArg() > 0 {
		config.Path = flag.Arg(0)
	}

	if *noColor {
		config.UseColor = false
	}

	return config
}

func loadItems(remarkablePath string) (map[string]*Item, error) {
	metadataFiles, err := filepath.Glob(filepath.Join(remarkablePath, "*.metadata"))
	if err != nil {
		return nil, err
	}

	items := make(map[string]*Item)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Load PDF and EPUB files for type detection
	pdfFiles, _ := filepath.Glob(filepath.Join(remarkablePath, "*.pdf"))
	epubFiles, _ := filepath.Glob(filepath.Join(remarkablePath, "*.epub"))

	pdfMap := make(map[string]bool)
	epubMap := make(map[string]bool)

	for _, f := range pdfFiles {
		uuid := strings.TrimSuffix(filepath.Base(f), ".pdf")
		pdfMap[uuid] = true
	}

	for _, f := range epubFiles {
		uuid := strings.TrimSuffix(filepath.Base(f), ".epub")
		epubMap[uuid] = true
	}

	// Process metadata files concurrently
	for _, metadataFile := range metadataFiles {
		wg.Add(1)
		go func(file string) {
			defer wg.Done()

			uuid := strings.TrimSuffix(filepath.Base(file), ".metadata")

			data, err := os.ReadFile(file)
			if err != nil {
				return
			}

			var metadata Metadata
			if err := json.Unmarshal(data, &metadata); err != nil {
				return
			}

			if metadata.Deleted {
				return
			}

			if metadata.VisibleName == "" {
				metadata.VisibleName = "Unnamed"
			}
			if metadata.Type == "" {
				metadata.Type = "DocumentType"
			}

			item := &Item{
				UUID:   uuid,
				Name:   metadata.VisibleName,
				Type:   metadata.Type,
				Parent: metadata.Parent,
			}

			// Determine document type
			if metadata.Type != "CollectionType" {
				if epubMap[uuid] {
					item.DocType = "epub"
				} else if pdfMap[uuid] {
					item.DocType = "pdf"
				} else {
					item.DocType = "notebook"
				}
			}

			// Create sort key: 0 for folders, 1 for documents, then name
			sortPrefix := "1"
			if metadata.Type == "CollectionType" {
				sortPrefix = "0"
			}
			item.SortKey = sortPrefix + "|" + metadata.VisibleName

			mu.Lock()
			items[uuid] = item
			mu.Unlock()
		}(metadataFile)
	}

	wg.Wait()
	return items, nil
}

func buildChildrenMap(items map[string]*Item) map[string][]*Item {
	children := make(map[string][]*Item)

	for _, item := range items {
		parent := item.Parent
		if parent == "" {
			parent = "root"
		}
		children[parent] = append(children[parent], item)
	}

	return children
}

func sortItems(items map[string]*Item, children map[string][]*Item) {
	for parent := range children {
		sort.Slice(children[parent], func(i, j int) bool {
			return children[parent][i].SortKey < children[parent][j].SortKey
		})
	}
}

func printTree(items map[string]*Item, children map[string][]*Item, config Config) {
	fmt.Println(".")

	roots := children["root"]
	trashItems := children["trash"]

	dirCount := 0
	fileCount := 0

	for _, item := range items {
		if item.Type == "CollectionType" {
			dirCount++
		} else {
			fileCount++
		}
	}

	// Print root items
	for i, item := range roots {
		isLast := i == len(roots)-1 && len(trashItems) == 0
		printItem(item, "", isLast, 0, children, config)
	}

	// Print trash items
	if len(trashItems) > 0 {
		dirCount++ // Add trash folder to count

		connector := "└── "
		icon := ""
		if config.ShowIcons {
			icon = "📁 "
		}

		color := ""
		colorReset := ""
		if config.UseColor {
			color = colors["folder"]
			colorReset = colors["reset"]
		}

		fmt.Printf("%s%s%sTrash%s\n", connector, color, icon, colorReset)

		for i, item := range trashItems {
			isLast := i == len(trashItems)-1
			printTrashItem(item, "    ", isLast, 1, config)
		}
	}

	fmt.Println()

	// Print summary
	dirText := "directories"
	if dirCount == 1 {
		dirText = "directory"
	}

	fileText := "files"
	if fileCount == 1 {
		fileText = "file"
	}

	fmt.Printf("%d %s, %d %s\n", dirCount, dirText, fileCount, fileText)
}

func printItem(item *Item, prefix string, isLast bool, depth int, children map[string][]*Item, config Config) {
	if depth > 50 {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	icon, color, typeLabel, uuidDisplay := getItemFormatting(item, config)

	fmt.Printf("%s%s%s%s%s%s%s%s\n", prefix, connector, color, icon, item.Name, colors["reset"], typeLabel, uuidDisplay)

	// Print children
	itemChildren := children[item.UUID]
	for i, child := range itemChildren {
		childIsLast := i == len(itemChildren)-1

		newPrefix := prefix
		if isLast {
			newPrefix += "    "
		} else {
			newPrefix += "│   "
		}

		printItem(child, newPrefix, childIsLast, depth+1, children, config)
	}
}

func printTrashItem(item *Item, prefix string, isLast bool, depth int, config Config) {
	if depth > 50 {
		return
	}

	connector := "├── "
	if isLast {
		connector = "└── "
	}

	icon, color, typeLabel, uuidDisplay := getItemFormatting(item, config)

	fmt.Printf("%s%s%s%s%s%s%s%s\n", prefix, connector, color, icon, item.Name, colors["reset"], typeLabel, uuidDisplay)
}

func getItemFormatting(item *Item, config Config) (icon, color, typeLabel, uuidDisplay string) {
	if config.UseColor {
		if item.Type == "CollectionType" {
			color = colors["folder"]
		} else {
			switch item.DocType {
			case "pdf":
				color = colors["pdf"]
			case "epub":
				color = colors["epub"]
			}
		}
	}

	if config.ShowIcons {
		if item.Type == "CollectionType" {
			icon = "📁 "
		} else {
			switch item.DocType {
			case "pdf":
				icon = "📕 "
			case "epub":
				icon = "📗 "
			default:
				icon = "📓 "
			}
		}
	}

	if config.ShowLabels && item.Type != "CollectionType" {
		switch item.DocType {
		case "pdf":
			typeLabel = " (pdf)"
		case "epub":
			typeLabel = " (epub)"
		default:
			typeLabel = " (notebook)"
		}
	}

	if config.ShowUUID && item.Type != "CollectionType" {
		uuidDisplay = " [" + item.UUID + "]"
	}

	return
}