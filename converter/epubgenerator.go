// Package converter provides FB2 to EPUB conversion functionality.
package converter

import (
	"archive/zip"
	"encoding/base64"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lex/fb2epub/models"
)

const (
	defaultTitle  = "Untitled"
	defaultAuthor = "Unknown"
)

// GenerateEPUB creates an EPUB file from an FB2 book
func GenerateEPUB(fb2 *models.FictionBook, outputPath string) error {
	// Create output directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create EPUB file (which is a ZIP archive)
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create EPUB file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log error but don't fail if file was written successfully
			_ = closeErr
		}
	}()

	zipWriter := zip.NewWriter(file)
	defer func() {
		if closeErr := zipWriter.Close(); closeErr != nil {
			_ = closeErr
		}
	}()

	// Add mimetype file (must be first, uncompressed)
	if err := addMimetype(zipWriter); err != nil {
		return err
	}

	// Add META-INF/container.xml
	if err := addContainer(zipWriter); err != nil {
		return err
	}

	// Collect images first (needed for manifest)
	imageMap := collectImages(fb2)

	// Add OEBPS/content.opf (package document)
	if err := addContentOPF(zipWriter, fb2, imageMap); err != nil {
		return err
	}

	// Add OEBPS/toc.ncx (navigation)
	if err := addTOCNCX(zipWriter, fb2); err != nil {
		return err
	}

	// Add EPUB 3.0 nav document
	if err := addNavXHTML(zipWriter, fb2); err != nil {
		return err
	}

	// Add HTML content files (need imageMap for image references)
	if err := addHTMLContent(zipWriter, fb2, imageMap); err != nil {
		return err
	}

	// Add binary resources (images)
	if err := addBinaryResources(zipWriter, fb2, imageMap); err != nil {
		return err
	}

	return nil
}

func addMimetype(writer *zip.Writer) error {
	header := &zip.FileHeader{
		Name:   "mimetype",
		Method: zip.Store, // Must be stored, not compressed
	}
	header.SetMode(0644)

	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte("application/epub+zip"))
	return err
}

func addContainer(writer *zip.Writer) error {
	w, err := writer.Create("META-INF/container.xml")
	if err != nil {
		return err
	}

	content := `<?xml version="1.0" encoding="UTF-8"?>
<container version="1.0" xmlns="urn:oasis:names:tc:opendocument:xmlns:container">
  <rootfiles>
    <rootfile full-path="OEBPS/content.opf" media-type="application/oebps-package+xml"/>
  </rootfiles>
</container>`

	_, err = w.Write([]byte(content))
	return err
}

func addContentOPF(writer *zip.Writer, fb2 *models.FictionBook, imageMap map[string]*ImageInfo) error {
	w, err := writer.Create("OEBPS/content.opf")
	if err != nil {
		return err
	}

	// Extract metadata
	title := fb2.Description.TitleInfo.BookTitle
	if title == "" {
		title = defaultTitle
	}

	authors := make([]string, 0)
	for _, author := range fb2.Description.TitleInfo.Author {
		name := buildAuthorName(author)
		if name != "" {
			authors = append(authors, name)
		}
	}
	authorStr := strings.Join(authors, ", ")
	if authorStr == "" {
		authorStr = defaultAuthor
	}

	lang := fb2.Description.TitleInfo.Lang
	if lang == "" {
		lang = "en"
	}

	uuid := "urn:uuid:" + generateUUID()
	date := time.Now().Format("2006-01-02")

	// Build manifest items
	manifestItems := `<item id="ncx" href="toc.ncx" media-type="application/x-dtbncx+xml" properties="nav"/>
    <item id="nav" href="nav.xhtml" media-type="application/xhtml+xml" properties="nav"/>
    <item id="cover" href="cover.xhtml" media-type="application/xhtml+xml"/>
    <item id="content" href="content.xhtml" media-type="application/xhtml+xml"/>`

	// Add image items to manifest
	for imgID, imgInfo := range imageMap {
		ext := getImageExtension(imgInfo.ContentType)
		manifestItems += fmt.Sprintf("\n    <item id=\"%s\" href=\"images/%s%s\" "+
			"media-type=\"%s\"/>", imgID, imgID, ext, imgInfo.ContentType)
	}

	// Build spine
	spine := `<itemref idref="cover"/>
    <itemref idref="content"/>`

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<package xmlns="http://www.idpf.org/2007/opf" version="3.0" unique-identifier="bookid">
  <metadata xmlns:dc="http://purl.org/dc/elements/1.1/">
    <dc:title>%s</dc:title>
    <dc:creator>%s</dc:creator>
    <dc:language>%s</dc:language>
    <dc:identifier id="bookid">%s</dc:identifier>
    <meta property="dcterms:modified">%s</meta>
  </metadata>
  <manifest>
    %s
  </manifest>
  <spine toc="ncx">
    %s
  </spine>
</package>`, html.EscapeString(title), html.EscapeString(authorStr), lang, uuid, date, manifestItems, spine)

	_, err = w.Write([]byte(content))
	return err
}

// TOCEntry represents a table of contents entry
type TOCEntry struct {
	ID        string
	Title     string
	PlayOrder int
	Children  []*TOCEntry
}

func addTOCNCX(writer *zip.Writer, fb2 *models.FictionBook) error {
	w, err := writer.Create("OEBPS/toc.ncx")
	if err != nil {
		return err
	}

	title := fb2.Description.TitleInfo.BookTitle
	if title == "" {
		title = "Untitled"
	}

	uuid := "urn:uuid:" + generateUUID()

	// Build TOC from sections
	tocEntries := buildTOC(fb2)

	// Calculate depth
	maxDepth := calculateTOCDepth(tocEntries, 0)
	if maxDepth < 1 {
		maxDepth = 1
	}

	// Build navMap XML
	var navMap strings.Builder
	playOrder := 1

	// Add cover
	navMap.WriteString(fmt.Sprintf(`    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel>
        <text>Cover</text>
      </navLabel>
      <content src="cover.xhtml"/>
    </navPoint>
`, playOrder, playOrder))
	playOrder++

	// Add content entry
	navMap.WriteString(fmt.Sprintf(`    <navPoint id="navpoint-%d" playOrder="%d">
      <navLabel>
        <text>Content</text>
      </navLabel>
      <content src="content.xhtml"/>
    </navPoint>
`, playOrder, playOrder))
	playOrder++

	// Add all section entries
	for _, entry := range tocEntries {
		playOrder = writeTOCEntry(&navMap, entry, playOrder, 0)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<ncx xmlns="http://www.daisy.org/z3986/2005/ncx/" version="2005-1">
  <head>
    <meta name="dtb:uid" content="%s"/>
    <meta name="dtb:depth" content="%d"/>
    <meta name="dtb:totalPageCount" content="0"/>
    <meta name="dtb:maxPageNumber" content="0"/>
  </head>
  <docTitle>
    <text>%s</text>
  </docTitle>
  <navMap>
%s  </navMap>
</ncx>`, uuid, maxDepth+1, html.EscapeString(title), navMap.String())

	_, err = w.Write([]byte(content))
	return err
}

func buildTOC(fb2 *models.FictionBook) []*TOCEntry {
	var entries []*TOCEntry

	// Process body sections
	for i, section := range fb2.Body.Section {
		if entry := buildTOCFromSection(&section, fmt.Sprintf("section-%d", i)); entry != nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func buildTOCFromSection(section *models.Section, baseID string) *TOCEntry {
	// Only create TOC entry if section has a title
	if section.Title == nil || len(section.Title.Paragraph) == 0 {
		// If no title but has subsections, still process children
		var children []*TOCEntry
		for i, subSection := range section.Section {
			if child := buildTOCFromSection(&subSection, fmt.Sprintf("%s-sub-%d", baseID, i)); child != nil {
				children = append(children, child)
			}
		}
		if len(children) > 0 {
			return &TOCEntry{
				ID:       baseID,
				Title:    "", // No title
				Children: children,
			}
		}
		return nil
	}

	// Extract title text
	var titleParts []string
	for _, p := range section.Title.Paragraph {
		text := processParagraph(&p, nil) // TOC doesn't need images
		if text != "" {
			titleParts = append(titleParts, text)
		}
	}
	title := strings.Join(titleParts, " ")
	if title == "" {
		title = "Untitled Section"
	}

	// Process children
	var children []*TOCEntry
	for i, subSection := range section.Section {
		if child := buildTOCFromSection(&subSection, fmt.Sprintf("%s-sub-%d", baseID, i)); child != nil {
			children = append(children, child)
		}
	}

	return &TOCEntry{
		ID:       baseID,
		Title:    title,
		Children: children,
	}
}

func writeTOCEntry(builder *strings.Builder, entry *TOCEntry, playOrder int, indent int) int {
	if entry.Title == "" && len(entry.Children) == 0 {
		return playOrder
	}

	indentStr := strings.Repeat("    ", indent+2)
	currentOrder := playOrder

	if entry.Title != "" {
		escapedTitle := html.EscapeString(entry.Title)
		fmt.Fprintf(builder, `%s<navPoint id="navpoint-%s" playOrder="%d">
%s  <navLabel>
%s    <text>%s</text>
%s  </navLabel>
%s  <content src="content.xhtml#%s"/>
`, indentStr, entry.ID, currentOrder, indentStr, indentStr, escapedTitle, indentStr, indentStr, entry.ID)

		currentOrder++

		// Process children
		if len(entry.Children) > 0 {
			for _, child := range entry.Children {
				currentOrder = writeTOCEntry(builder, child, currentOrder, indent+1)
			}
		}

		fmt.Fprintf(builder, "%s</navPoint>\n", indentStr)
		return currentOrder
	}
	// No title, just process children
	for _, child := range entry.Children {
		currentOrder = writeTOCEntry(builder, child, currentOrder, indent)
	}
	return currentOrder
}

func calculateTOCDepth(entries []*TOCEntry, currentDepth int) int {
	maxDepth := currentDepth
	for _, entry := range entries {
		depth := calculateTOCDepth(entry.Children, currentDepth+1)
		if depth > maxDepth {
			maxDepth = depth
		}
	}
	return maxDepth
}

func addHTMLContent(writer *zip.Writer, fb2 *models.FictionBook, imageMap map[string]*ImageInfo) error {
	// Add cover page
	if err := addCoverPage(writer, fb2, imageMap); err != nil {
		return err
	}

	// Add main content
	if err := addMainContent(writer, fb2, imageMap); err != nil {
		return err
	}

	return nil
}

func addCoverPage(writer *zip.Writer, fb2 *models.FictionBook, _ map[string]*ImageInfo) error {
	w, err := writer.Create("OEBPS/cover.xhtml")
	if err != nil {
		return err
	}

	title := fb2.Description.TitleInfo.BookTitle
	if title == "" {
		title = "Untitled"
	}

	authors := make([]string, 0)
	for _, author := range fb2.Description.TitleInfo.Author {
		name := buildAuthorName(author)
		if name != "" {
			authors = append(authors, name)
		}
	}
	authorStr := strings.Join(authors, ", ")
	if authorStr == "" {
		authorStr = defaultAuthor
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>%s</title>
  <style type="text/css">
    body { text-align: center; padding: 2em; font-family: serif; }
    h1 { margin-top: 3em; }
    h2 { margin-top: 2em; color: #666; }
  </style>
</head>
<body>
  <h1>%s</h1>
  <h2>%s</h2>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), html.EscapeString(authorStr))

	_, err = w.Write([]byte(content))
	return err
}

func addMainContent(writer *zip.Writer, fb2 *models.FictionBook, imageMap map[string]*ImageInfo) error {
	w, err := writer.Create("OEBPS/content.xhtml")
	if err != nil {
		return err
	}

	var bodyContent strings.Builder
	bodyContent.WriteString(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Content</title>
  <style type="text/css">
    body { font-family: serif; padding: 1em; line-height: 1.6; }
    h1, h2, h3 { margin-top: 1.5em; }
    p { margin: 1em 0; text-align: justify; }
    .empty-line { height: 1em; }
    strong { font-weight: bold; }
    em { font-style: italic; }
  </style>
</head>
<body>
`)

	// Process body title if present
	if len(fb2.Body.Title.Paragraph) > 0 {
		for _, p := range fb2.Body.Title.Paragraph {
			text := processParagraph(&p, imageMap)
			bodyContent.WriteString(fmt.Sprintf("<h1>%s</h1>\n", text))
		}
	}

	// Process body sections
	for i := range fb2.Body.Section {
		processSectionWithID(&bodyContent, &fb2.Body.Section[i], 0, i, "", imageMap)
	}

	bodyContent.WriteString(`</body>
</html>`)

	_, err = w.Write([]byte(bodyContent.String()))
	return err
}

func processSectionWithID(
	builder *strings.Builder,
	section *models.Section,
	depth int,
	sectionIndex int,
	parentID string,
	imageMap map[string]*ImageInfo,
) {
	sectionID := ""
	if parentID != "" {
		sectionID = fmt.Sprintf("%s-sub-%d", parentID, sectionIndex)
	} else {
		sectionID = fmt.Sprintf("section-%d", sectionIndex)
	}

	// Add title if present
	if section.Title != nil && len(section.Title.Paragraph) > 0 {
		level := depth + 1
		if level > 6 {
			level = 6
		}
		tag := fmt.Sprintf("h%d", level)
		for _, p := range section.Title.Paragraph {
			text := processParagraph(&p, nil) // Titles don't need images
			// Ensure sectionID is safe for XML (no special characters)
			safeID := html.EscapeString(sectionID)
			fmt.Fprintf(builder, "<%s id=\"%s\">%s</%s>\n", tag, safeID, text, tag)
		}
	}

	// Add paragraphs
	for _, p := range section.Paragraph {
		text := processParagraph(&p, imageMap)
		if text != "" {
			fmt.Fprintf(builder, "<p>%s</p>\n", text)
		}
	}

	// Add empty lines
	for range section.EmptyLine {
		builder.WriteString(`<div class="empty-line"></div>` + "\n")
	}

	// Process nested sections
	for i := range section.Section {
		processSectionWithID(builder, &section.Section[i], depth+1, i, sectionID, imageMap)
	}

	// Process poems
	for _, poem := range section.Poem {
		processPoem(builder, &poem)
	}

	// Process citations
	for _, cite := range section.Cite {
		processCite(builder, &cite, imageMap)
	}
}

// processParagraph processes a paragraph and preserves all text attributes
func processParagraph(p *models.Paragraph, imageMap map[string]*ImageInfo) string {
	var result strings.Builder

	// Start with base text
	if p.Text != "" {
		result.WriteString(html.EscapeString(p.Text))
	}

	// Process inline elements in order
	// Note: Go's XML unmarshaling doesn't preserve exact order of mixed content,
	// but we process elements to preserve their attributes

	// Process links first (they might be nested in strong/emphasis)
	for _, link := range p.Link {
		linkHTML := processLink(&link, imageMap)
		// Try to find and replace the link text in the paragraph text
		if link.Text != "" {
			escapedLinkText := html.EscapeString(link.Text)
			current := result.String()
			if strings.Contains(current, escapedLinkText) {
				// Replace the text with the link HTML
				result.Reset()
				result.WriteString(strings.Replace(current, escapedLinkText, linkHTML, 1))
			} else {
				// If not found, append it
				result.WriteString(" " + linkHTML)
			}
		} else {
			result.WriteString(" " + linkHTML)
		}
	}

	// Process strong elements (may contain nested elements)
	for _, strong := range p.Strong {
		strongHTML := processStrong(&strong, imageMap)
		// Try to find and replace
		if strong.Text != "" || len(strong.Link) > 0 {
			strongText := extractStrongText(&strong)
			if strongText != "" {
				escapedStrongText := html.EscapeString(strongText)
				current := result.String()
				if strings.Contains(current, escapedStrongText) {
					result.Reset()
					result.WriteString(strings.Replace(current, escapedStrongText, strongHTML, 1))
				} else {
					result.WriteString(" " + strongHTML)
				}
			} else {
				result.WriteString(" " + strongHTML)
			}
		} else {
			result.WriteString(" " + strongHTML)
		}
	}

	// Process emphasis elements (may contain nested elements)
	for _, emphasis := range p.Emphasis {
		emphasisHTML := processEmphasis(&emphasis, imageMap)
		// Try to find and replace
		if emphasis.Text != "" || len(emphasis.Link) > 0 {
			emphasisText := extractEmphasisText(&emphasis)
			if emphasisText != "" {
				escapedEmphasisText := html.EscapeString(emphasisText)
				current := result.String()
				if strings.Contains(current, escapedEmphasisText) {
					result.Reset()
					result.WriteString(strings.Replace(current, escapedEmphasisText, emphasisHTML, 1))
				} else {
					result.WriteString(" " + emphasisHTML)
				}
			} else {
				result.WriteString(" " + emphasisHTML)
			}
		} else {
			result.WriteString(" " + emphasisHTML)
		}
	}

	// Process images - insert inline
	for _, image := range p.Image {
		href := html.EscapeString(image.Href)
		imgID := strings.TrimPrefix(href, "#")

		var imgPath string
		if imageMap != nil {
			if imgInfo, exists := imageMap[imgID]; exists {
				ext := getImageExtension(imgInfo.ContentType)
				imgPath = fmt.Sprintf("images/%s%s", imgID, ext)
			} else {
				imgPath = fmt.Sprintf("images/%s.jpg", imgID)
			}
		} else {
			imgPath = fmt.Sprintf("images/%s.jpg", imgID)
		}
		result.WriteString(fmt.Sprintf(" <img src=\"%s\" alt=\"\"/>", html.EscapeString(imgPath)))
	}

	return result.String()
}

// processStrong processes a strong element and its nested content
func processStrong(s *models.Strong, imageMap map[string]*ImageInfo) string {
	var result strings.Builder

	if s.Text != "" {
		result.WriteString(html.EscapeString(s.Text))
	}

	// Process nested links
	for _, link := range s.Link {
		linkHTML := processLink(&link, imageMap)
		if s.Text != "" && link.Text != "" {
			escapedLinkText := html.EscapeString(link.Text)
			current := result.String()
			if strings.Contains(current, escapedLinkText) {
				result.Reset()
				result.WriteString(strings.Replace(current, escapedLinkText, linkHTML, 1))
			} else {
				result.WriteString(" " + linkHTML)
			}
		} else {
			result.WriteString(linkHTML)
		}
	}

	// Process nested emphasis
	for _, emphasis := range s.Emphasis {
		emphasisHTML := processEmphasis(&emphasis, imageMap)
		result.WriteString(emphasisHTML)
	}

	// Process nested strong
	for _, nestedStrong := range s.Strong {
		nestedHTML := processStrong(&nestedStrong, imageMap)
		result.WriteString(nestedHTML)
	}

	return "<strong>" + result.String() + "</strong>"
}

// processEmphasis processes an emphasis element and its nested content
func processEmphasis(e *models.Emphasis, imageMap map[string]*ImageInfo) string {
	var result strings.Builder

	if e.Text != "" {
		result.WriteString(html.EscapeString(e.Text))
	}

	// Process nested links
	for _, link := range e.Link {
		linkHTML := processLink(&link, imageMap)
		if e.Text != "" && link.Text != "" {
			escapedLinkText := html.EscapeString(link.Text)
			current := result.String()
			if strings.Contains(current, escapedLinkText) {
				result.Reset()
				result.WriteString(strings.Replace(current, escapedLinkText, linkHTML, 1))
			} else {
				result.WriteString(" " + linkHTML)
			}
		} else {
			result.WriteString(linkHTML)
		}
	}

	// Process nested strong
	for _, strong := range e.Strong {
		strongHTML := processStrong(&strong, imageMap)
		result.WriteString(strongHTML)
	}

	// Process nested emphasis
	for _, nestedEmphasis := range e.Emphasis {
		nestedHTML := processEmphasis(&nestedEmphasis, imageMap)
		result.WriteString(nestedHTML)
	}

	return "<em>" + result.String() + "</em>"
}

// processLink processes a link element
func processLink(l *models.Link, _ map[string]*ImageInfo) string {
	href := html.EscapeString(l.Href)
	text := html.EscapeString(l.Text)
	if text == "" {
		text = href // Use href as text if no text provided
	}
	return fmt.Sprintf("<a href=\"%s\">%s</a>", href, text)
}

// extractStrongText extracts the text content from a strong element
func extractStrongText(s *models.Strong) string {
	var result strings.Builder
	result.WriteString(s.Text)
	for _, link := range s.Link {
		result.WriteString(link.Text)
	}
	for _, emphasis := range s.Emphasis {
		result.WriteString(extractEmphasisText(&emphasis))
	}
	for _, nestedStrong := range s.Strong {
		result.WriteString(extractStrongText(&nestedStrong))
	}
	return result.String()
}

// extractEmphasisText extracts the text content from an emphasis element
func extractEmphasisText(e *models.Emphasis) string {
	var result strings.Builder
	result.WriteString(e.Text)
	for _, link := range e.Link {
		result.WriteString(link.Text)
	}
	for _, strong := range e.Strong {
		result.WriteString(extractStrongText(&strong))
	}
	for _, nestedEmphasis := range e.Emphasis {
		result.WriteString(extractEmphasisText(&nestedEmphasis))
	}
	return result.String()
}

func processPoem(builder *strings.Builder, poem *models.Poem) {
	builder.WriteString("<div class=\"poem\">\n")

	if poem.Title != nil {
		builder.WriteString("<h3>")
		for _, p := range poem.Title.Paragraph {
			builder.WriteString(html.EscapeString(p.Text))
		}
		builder.WriteString("</h3>\n")
	}

	for _, stanza := range poem.Stanza {
		builder.WriteString("<div class=\"stanza\">\n")
		for _, verse := range stanza.Verse {
			fmt.Fprintf(builder, "<p class=\"verse\">%s</p>\n", html.EscapeString(verse.Text))
		}
		builder.WriteString("</div>\n")
	}

	builder.WriteString("</div>\n")
}

func processCite(builder *strings.Builder, cite *models.Cite, imageMap map[string]*ImageInfo) {
	builder.WriteString("<blockquote class=\"cite\">\n")
	for _, p := range cite.Paragraph {
		text := processParagraph(&p, imageMap)
		fmt.Fprintf(builder, "<p>%s</p>\n", text)
	}
	builder.WriteString("</blockquote>\n")
}

// ImageInfo stores image metadata
type ImageInfo struct {
	ContentType string
	Data        []byte
}

func collectImages(fb2 *models.FictionBook) map[string]*ImageInfo {
	imageMap := make(map[string]*ImageInfo)
	for _, binary := range fb2.Binary {
		// Decode base64 data
		data, err := base64.StdEncoding.DecodeString(binary.Data)
		if err != nil {
			// Skip invalid base64 data
			continue
		}
		imageMap[binary.ID] = &ImageInfo{
			ContentType: binary.ContentType,
			Data:        data,
		}
	}
	return imageMap
}

func getImageExtension(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "image/svg+xml":
		return ".svg"
	default:
		return ".jpg" // Default to jpg
	}
}

func addBinaryResources(writer *zip.Writer, _ *models.FictionBook, imageMap map[string]*ImageInfo) error {
	// Create images directory entry
	for imgID, imgInfo := range imageMap {
		ext := getImageExtension(imgInfo.ContentType)
		path := fmt.Sprintf("OEBPS/images/%s%s", imgID, ext)

		w, err := writer.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create image file %s: %w", path, err)
		}

		_, err = w.Write(imgInfo.Data)
		if err != nil {
			return fmt.Errorf("failed to write image data %s: %w", path, err)
		}
	}
	return nil
}

func buildAuthorName(author models.Author) string {
	parts := make([]string, 0)
	if author.FirstName != "" {
		parts = append(parts, author.FirstName)
	}
	if author.MiddleName != "" {
		parts = append(parts, author.MiddleName)
	}
	if author.LastName != "" {
		parts = append(parts, author.LastName)
	}
	if len(parts) == 0 && author.Nickname != "" {
		return author.Nickname
	}
	return strings.Join(parts, " ")
}

func generateUUID() string {
	return uuid.New().String()
}
