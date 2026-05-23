package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// copied from <https://github.com/docker/cli-docs-tool/blob/7543983fe39fcedd9182dfbee8824375999f8add/clidocstool_yaml.go#L32>.
type cmdOption struct {
	Option          string
	Shorthand       string `yaml:",omitempty"`
	ValueType       string `yaml:"value_type,omitempty"`
	DefaultValue    string `yaml:"default_value,omitempty"`
	Description     string `yaml:",omitempty"`
	DetailsURL      string `yaml:"details_url,omitempty"` // DetailsURL contains an anchor-id or link for more information on this flag
	Deprecated      bool
	Hidden          bool
	MinAPIVersion   string `yaml:"min_api_version,omitempty"`
	Experimental    bool
	ExperimentalCLI bool
	Kubernetes      bool
	Swarm           bool
	OSType          string `yaml:"os_type,omitempty"`
}

// copied from <https://github.com/docker/cli-docs-tool/blob/7543983fe39fcedd9182dfbee8824375999f8add/clidocstool_yaml.go#L49>.
type cmdDoc struct {
	Name             string      `yaml:"command"`
	SeeAlso          []string    `yaml:"parent,omitempty"`
	Version          string      `yaml:"engine_version,omitempty"`
	Aliases          string      `yaml:",omitempty"`
	Short            string      `yaml:",omitempty"`
	Long             string      `yaml:",omitempty"`
	Usage            string      `yaml:",omitempty"`
	Pname            string      `yaml:",omitempty"`
	Plink            string      `yaml:",omitempty"`
	Cname            []string    `yaml:",omitempty"`
	Clink            []string    `yaml:",omitempty"`
	Options          []cmdOption `yaml:",omitempty"`
	InheritedOptions []cmdOption `yaml:"inherited_options,omitempty"`
	Example          string      `yaml:"examples,omitempty"`
	Deprecated       bool
	Hidden           bool
	MinAPIVersion    string `yaml:"min_api_version,omitempty"`
	Experimental     bool
	ExperimentalCLI  bool
	Kubernetes       bool
	Swarm            bool
	OSType           string `yaml:"os_type,omitempty"`
}

// outputPathFor generates the output path for a given command.
//
// Examples:
//   - "wpm" => "index.mdx"
//   - "wpm install" => "install.mdx"
//   - "wpm auth" => "auth/index.mdx"
//   - "wpm auth login" => "auth/login.mdx"
func outputPathFor(command string, hasChildren bool) string {
	segments := strings.Split(command, " ")[1:]
	if len(segments) == 0 {
		return "index.mdx"
	}
	base := strings.Join(segments, "/")
	if hasChildren {
		return fmt.Sprintf("%s/index.mdx", base)
	}
	return fmt.Sprintf("%s.mdx", base)
}

var alertTypeMap = map[string]string{
	"NOTE":      "info",
	"TIP":       "info",
	"IMPORTANT": "info",
	"WARNING":   "warn",
	"CAUTION":   "error",
}

// titleCase converts a word to title case.
//
// Examples:
//   - "install" => "Install"
func titleCase(word string) string {
	if len(word) == 0 {
		return ""
	}
	return strings.ToUpper(string(word[0])) + strings.ToLower(word[1:])
}

// buildMarkdownTable constructs a markdown table given headers and rows.
func buildMarkdownTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.Grow(512)

	// Add header.
	sb.WriteString("|")
	for _, h := range headers {
		fmt.Fprintf(&sb, " %s |", h)
	}
	sb.WriteString("\n|")
	for range headers {
		sb.WriteString(" --- |")
	}
	sb.WriteString("\n")

	// Add rows.
	for _, row := range rows {
		sb.WriteString("|")
		for _, cell := range row {
			cell = strings.ReplaceAll(cell, "|", "\\|")
			cell = strings.ReplaceAll(cell, "\n", " ")

			fmt.Fprintf(&sb, " %s |", cell)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

var blockquotePrefix = regexp.MustCompile(`^[ \t]*>[ \t]?`)

// getBlockquoteLineBounds determines the absolute start and end byte offsets
// of a blockquote node in the original source string.
func getBlockquoteLineBounds(src []byte, n *ast.Blockquote) (int, int) {
	minStart := len(src)
	maxStop := -1

	ast.Walk(n, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && child.Type() == ast.TypeBlock {
			lines := child.Lines()
			if lines != nil {
				for i := 0; i < lines.Len(); i++ {
					seg := lines.At(i)
					if seg.Start < minStart {
						minStart = seg.Start
					}
					if seg.Stop > maxStop {
						maxStop = seg.Stop
					}
				}
			}
		}
		return ast.WalkContinue, nil
	})

	if minStart > maxStop {
		return -1, -1
	}

	start := minStart
	for start > 0 && src[start-1] != '\n' {
		start--
	}
	end := maxStop
	for end < len(src) && src[end] != '\n' {
		end++
	}
	if end < len(src) && src[end] == '\n' {
		end++
	}

	return start, end
}

// processAlertBlock safely strips the outer blockquote prefix while preserving
// internal formatting (like nested quotes or code blocks).
func processAlertBlock(block string, keyword string) string {
	lines := strings.Split(block, "\n")
	var content []string

	for i, line := range lines {
		if i == len(lines)-1 && line == "" {
			continue
		}

		cleanLine := blockquotePrefix.ReplaceAllString(line, "")

		if i == 0 {
			_, after, ok := strings.Cut(cleanLine, "]")
			if ok && strings.HasPrefix(strings.TrimSpace(cleanLine), "[!") {
				remainder := strings.TrimSpace(after)
				if remainder != "" {
					content = append(content, remainder)
				}
			} else {
				content = append(content, cleanLine)
			}
		} else {
			content = append(content, cleanLine)
		}
	}

	return fmt.Sprintf(
		"\n\n<Callout type=\"%s\" title=\"%s\">\n%s\n</Callout>\n\n",
		alertTypeMap[keyword],
		titleCase(keyword),
		strings.TrimSpace(strings.Join(content, "\n")),
	)
}

var (
	parser    = goldmark.DefaultParser()
	commentRE = regexp.MustCompile(`(?s)<!--.*?(-->|$)`)
)

// stripHTMLComments removes HTML comments from the input string.
func stripHTMLComments(src string) string {
	if !strings.Contains(src, "<!--") {
		return src
	}

	source := []byte(src)
	doc := parser.Parse(text.NewReader(source))

	type replacement struct{ start, end int }
	var replacements []replacement

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering && n.Type() == ast.TypeBlock && n.Kind() == ast.KindHTMLBlock {
			minStart, maxStop := len(source), -1
			lines := n.Lines()
			if lines != nil {
				for i := 0; i < lines.Len(); i++ {
					seg := lines.At(i)
					if seg.Start < minStart {
						minStart = seg.Start
					}
					if seg.Stop > maxStop {
						maxStop = seg.Stop
					}
				}
			}
			if minStart < maxStop {
				blockText := string(source[minStart:maxStop])
				if strings.HasPrefix(strings.TrimSpace(blockText), "<!--") {
					replacements = append(replacements, replacement{minStart, maxStop})
				}
			}
		}
		return ast.WalkContinue, nil
	})

	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start > replacements[j].start
	})

	res := source
	for _, r := range replacements {
		end := r.end
		if end < len(res) && res[end] == '\n' {
			end++
		}
		prefix := res[:r.start]
		suffix := res[end:]

		newRes := make([]byte, 0, len(prefix)+len(suffix))
		newRes = append(newRes, prefix...)
		newRes = append(newRes, suffix...)
		res = newRes
	}

	return commentRE.ReplaceAllString(string(res), "")
}

type replacement struct {
	start, end int
	newText    string
}

// convertAlerts replaces GFM "Note"/"Warning" highlights with MDX `<Callout>` blocks.
func convertAlerts(src string) string {
	source := []byte(src)
	doc := parser.Parse(text.NewReader(source))

	var replacements []replacement

	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		bq, ok := n.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}

		var para *ast.Paragraph
		for c := bq.FirstChild(); c != nil; c = c.NextSibling() {
			if p, ok := c.(*ast.Paragraph); ok {
				para = p
				break
			}
		}

		if para == nil || para.Lines().Len() == 0 {
			return ast.WalkContinue, nil
		}

		seg := para.Lines().At(0)
		firstLine := string(seg.Value(source))

		keyword := ""
		for k := range alertTypeMap {
			if strings.HasPrefix(strings.TrimSpace(firstLine), "[!"+k+"]") {
				keyword = k
				break
			}
		}

		if keyword == "" {
			return ast.WalkContinue, nil
		}

		start, end := getBlockquoteLineBounds(source, bq)
		if start == -1 {
			return ast.WalkContinue, nil
		}

		blockStr := string(source[start:end])
		newText := processAlertBlock(blockStr, keyword)

		replacements = append(replacements, replacement{
			start:   start,
			end:     end,
			newText: newText,
		})

		return ast.WalkSkipChildren, nil
	})

	sort.Slice(replacements, func(i, j int) bool {
		return replacements[i].start < replacements[j].start
	})

	var out strings.Builder
	out.Grow(len(source) + 512)

	lastEnd := 0
	for _, r := range replacements {
		if r.start < lastEnd {
			continue
		}
		out.Write(source[lastEnd:r.start])
		out.WriteString(r.newText)
		lastEnd = r.end
	}
	if lastEnd < len(source) {
		out.Write(source[lastEnd:])
	}

	return out.String()
}

// processMarkdownText applies all necessary transformations to the markdown text.
func processMarkdownText(src string) string {
	if src == "" {
		return ""
	}

	clean := stripHTMLComments(src)
	clean = convertAlerts(clean)

	return strings.TrimSpace(clean)
}

var (
	multiNewlineRE    = regexp.MustCompile(`\n{3,}`)
	tableDescReplacer = strings.NewReplacer("|", "\\|", "\n", " ")

	// skipDefaultValues is a set of default values that should be skipped
	// when rendering the options table in the generated MDX files.
	skipDefaultValues = map[string]bool{
		"": true, "false": true, "[]": true, "0": true, "0s": true,
	}
)

type fm struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
}

// rendermarkdown generates the MDX content for a given cmdDoc, including frontmatter and alert processing.
func rendermarkdown(doc cmdDoc, byCommand map[string]cmdDoc) (string, error) {
	var sb strings.Builder
	sb.Grow(4096)

	fm, err := yaml.Marshal(fm{
		Title:       doc.Name,
		Description: doc.Short,
	})
	if err != nil {
		return "", err
	}

	sb.WriteString("---\n")
	sb.Write(fm)
	sb.WriteString("---\n\n")

	if doc.Deprecated {
		sb.WriteString(`<Callout type="warn" title="This command is deprecated">It may be removed in a future Docker version.</Callout>` + "\n\n")
	}

	if doc.Experimental || doc.ExperimentalCLI {
		sb.WriteString(`<Callout type="info" title="This command is experimental">Experimental features are intended for testing and feedback as their functionality or design may change between releases without warning or can be removed entirely in a future release.</Callout>` + "\n\n")
	}

	if doc.Long != "" {
		fmt.Fprintf(&sb, "## Description\n\n%s\n\n", processMarkdownText(doc.Long))
	}

	visibleOptions := []cmdOption{}
	for _, opt := range doc.Options {
		if opt.Hidden {
			continue
		}
		visibleOptions = append(visibleOptions, opt)
	}

	if len(visibleOptions) > 0 {
		sb.WriteString("## Options\n\n")

		var rows [][]string
		for _, opt := range visibleOptions {
			name := fmt.Sprintf("`--%s`", opt.Option)
			if opt.Shorthand != "" {
				name = fmt.Sprintf("`-%s, --%s`", opt.Shorthand, opt.Option)
			}

			defaultValue := ""
			if opt.DefaultValue != "" && !skipDefaultValues[opt.DefaultValue] {
				defaultValue = fmt.Sprintf("`%s`", opt.DefaultValue)
			}

			rows = append(rows, []string{name, defaultValue, opt.Description})
		}

		sb.WriteString(buildMarkdownTable([]string{"Option", "Default", "Description"}, rows))
		sb.WriteString("\n\n")
	}

	if doc.Example != "" {
		fmt.Fprintf(&sb, "## Examples\n\n%s\n\n", processMarkdownText(doc.Example))
	}

	if len(doc.Cname) > 0 {
		depth := len(strings.Split(doc.Name, " "))
		var rows [][]string
		for _, childCmd := range doc.Cname {
			child, exists := byCommand[childCmd]
			if !exists || child.Hidden {
				continue
			}

			relative := strings.Join(strings.Split(childCmd, " ")[depth:], "/")
			if len(child.Cname) > 0 {
				relative += "/index.mdx"
			} else {
				relative += ".mdx"
			}

			link := fmt.Sprintf("[`%s`](./%s)", childCmd, relative)
			rows = append(rows, []string{link, child.Short})
		}

		if len(rows) > 0 {
			sb.WriteString("## Subcommands\n\n")
			sb.WriteString(buildMarkdownTable([]string{"Command", "Description"}, rows))
			sb.WriteString("\n\n")
		}
	}

	result := multiNewlineRE.ReplaceAllString(sb.String(), "\n\n")
	return strings.TrimRight(result, "\n") + "\n", nil
}

// cleanMdxRecursively safely deletes ONLY `.mdx` files and empty folders.
// It preserves `meta.json` or other layout files you might need for Fumadocs.
func cleanMdxRecursively(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		full := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			if err := cleanMdxRecursively(full); err != nil {
				return err
			}
			os.Remove(full) // Only removes the folder if it is now empty
		} else if strings.HasSuffix(entry.Name(), ".mdx") {
			os.Remove(full)
		}
	}
	return nil
}

type options struct {
	inputDir  string
	outputDir string
}

type renderedPage struct {
	path    string
	content string
}

func main() {
	var opts options

	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.DateTime,
	}).With().Timestamp().Logger()

	cmd := &cobra.Command{
		Use:           "yml-to-mdx",
		Short:         "Convert YAML files to MDX format",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries, err := os.ReadDir(opts.inputDir)
			if err != nil {
				return err
			}

			var yamlFiles []string
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".yaml") {
					yamlFiles = append(yamlFiles, entry.Name())
				}
			}
			sort.Strings(yamlFiles)

			docs := make([]cmdDoc, 0, len(yamlFiles))
			byCommand := make(map[string]cmdDoc)

			for _, yamlFile := range yamlFiles {
				fullPath := filepath.Join(opts.inputDir, yamlFile)

				data, err := os.ReadFile(fullPath)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %w", fullPath, err)
				}

				var doc cmdDoc
				if err := yaml.Unmarshal(data, &doc); err != nil {
					return fmt.Errorf("failed to unmarshal YAML file %s: %w", fullPath, err)
				}
				docs = append(docs, doc)
				byCommand[doc.Name] = doc
			}

			var rendered []renderedPage
			for _, doc := range docs {
				if doc.Hidden {
					continue
				}

				md, err := rendermarkdown(doc, byCommand)
				if err != nil {
					return fmt.Errorf("failed to render markdown for command %s: %w", doc.Name, err)
				}

				outPath := filepath.Join(opts.outputDir, outputPathFor(doc.Name, len(doc.Cname) > 0))

				rendered = append(rendered, renderedPage{
					path:    outPath,
					content: md,
				})
			}

			// clean output directory
			if err := cleanMdxRecursively(opts.outputDir); err != nil {
				return fmt.Errorf("failed to safely clean output directory %s: %w", opts.outputDir, err)
			}
			if err := os.MkdirAll(opts.outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory %s: %w", opts.outputDir, err)
			}

			for _, page := range rendered {
				if err := os.MkdirAll(filepath.Dir(page.path), 0755); err != nil {
					return fmt.Errorf("failed to create directory for output file %s: %w", filepath.Dir(page.path), err)
				}

				if err := os.WriteFile(page.path, []byte(page.content), 0644); err != nil {
					return fmt.Errorf("failed to write output file %s: %w", page.path, err)
				}
			}

			logger.Info().Int("count", len(rendered)).Str("output_dir", opts.outputDir).Msg("Successfully converted YAML files to MDX format")

			return nil
		},
	}

	cmd.Flags().StringVarP(&opts.inputDir, "input-dir", "i", "", "Directory containing YAML files")
	cmd.Flags().StringVarP(&opts.outputDir, "output-dir", "o", "content/reference/cli", "Directory to save MDX files")

	if err := cmd.MarkFlagRequired("input-dir"); err != nil {
		logger.Fatal().Err(err).Msg("Failed to mark required flags")
	}

	if err := cmd.Execute(); err != nil {
		logger.Fatal().Err(err).Msg("YAML to MDX conversion failed")
	}
}
