package converter

import (
	"archive/zip"
	"fmt"
	"html"
	"strings"

	"github.com/lex/fb2epub/models"
)

// addNavXHTML creates EPUB 3.0 navigation document
// Note: defaultTitle is defined in epubgenerator.go
func addNavXHTML(writer *zip.Writer, fb2 *models.FictionBook) error {
	w, err := writer.Create("OEBPS/nav.xhtml")
	if err != nil {
		return err
	}

	title := fb2.Description.TitleInfo.BookTitle
	if title == "" {
		title = defaultTitle //nolint: ineffassign // title is used in fmt.Sprintf below
	}

	// Build TOC from sections
	tocEntries := buildTOC(fb2)

	// Build nav list
	var navList strings.Builder

	// Add cover
	navList.WriteString(`    <li><a href="cover.xhtml">Cover</a></li>
`)

	// Add content
	navList.WriteString(`    <li><a href="content.xhtml">Content</a></li>
`)

	// Add all section entries
	for _, entry := range tocEntries {
		writeNavEntry(&navList, entry, 0)
	}

	content := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html>
<html xmlns="http://www.w3.org/1999/xhtml" xmlns:epub="http://www.idpf.org/2007/ops">
<head>
  <title>Table of Contents</title>
  <style type="text/css">
    nav { font-family: serif; }
    ol { list-style-type: none; padding-left: 1em; }
    li { margin: 0.5em 0; }
    a { text-decoration: none; color: inherit; }
    a:hover { text-decoration: underline; }
  </style>
</head>
<body>
  <nav epub:type="toc" id="toc">
    <h1>Table of Contents</h1>
    <ol>
%s    </ol>
  </nav>
</body>
</html>`, navList.String())

	_, err = w.Write([]byte(content))
	return err
}

func writeNavEntry(builder *strings.Builder, entry *TOCEntry, indent int) {
	if entry.Title == "" && len(entry.Children) == 0 {
		return
	}

	indentStr := strings.Repeat("      ", indent+1)

	if entry.Title != "" {
		escapedID := html.EscapeString(entry.ID)
		escapedTitle := html.EscapeString(entry.Title)
		fmt.Fprintf(builder, `%s<li><a href="content.xhtml#%s">%s</a>`, indentStr, escapedID, escapedTitle)

		if len(entry.Children) > 0 {
			builder.WriteString("\n")
			fmt.Fprintf(builder, "%s  <ol>\n", indentStr)
			for _, child := range entry.Children {
				writeNavEntry(builder, child, indent+1)
			}
			fmt.Fprintf(builder, "%s  </ol>\n", indentStr)
			fmt.Fprintf(builder, "%s</li>\n", indentStr)
		} else {
			builder.WriteString("</li>\n")
		}
	} else {
		// No title, just process children
		for _, child := range entry.Children {
			writeNavEntry(builder, child, indent)
		}
	}
}
