// Package models provides data structures for FB2 (FictionBook 2.0) format.
package models

import "encoding/xml"

// FictionBook represents the root element of FB2 format
type FictionBook struct {
	XMLName     xml.Name    `xml:"FictionBook"`
	Description Description `xml:"description"`
	Body        Body        `xml:"body"`
	Binary      []Binary    `xml:"binary"`
}

// Description contains metadata about the book
type Description struct {
	TitleInfo    TitleInfo    `xml:"title-info"`
	PublishInfo  PublishInfo  `xml:"publish-info,omitempty"`
	DocumentInfo DocumentInfo `xml:"document-info,omitempty"`
}

// TitleInfo contains book title and author information
type TitleInfo struct {
	Genre      []string `xml:"genre"`
	Author     []Author `xml:"author"`
	BookTitle  string   `xml:"book-title"`
	Annotation string   `xml:"annotation,omitempty"`
	Date       string   `xml:"date,omitempty"`
	Lang       string   `xml:"lang,omitempty"`
}

// Author represents book author
type Author struct {
	FirstName  string `xml:"first-name,omitempty"`
	MiddleName string `xml:"middle-name,omitempty"`
	LastName   string `xml:"last-name,omitempty"`
	Nickname   string `xml:"nickname,omitempty"`
}

// PublishInfo contains publishing information
type PublishInfo struct {
	BookName  string `xml:"book-name,omitempty"`
	Publisher string `xml:"publisher,omitempty"`
	City      string `xml:"city,omitempty"`
	Year      string `xml:"year,omitempty"`
	ISBN      string `xml:"isbn,omitempty"`
}

// DocumentInfo contains document metadata
type DocumentInfo struct {
	Author      []Author `xml:"author,omitempty"`
	ProgramUsed string   `xml:"program-used,omitempty"`
	Date        string   `xml:"date,omitempty"`
	ID          string   `xml:"id,omitempty"`
	Version     string   `xml:"version,omitempty"`
}

// Body represents the main content of the book
type Body struct {
	Name    string    `xml:"name,attr,omitempty"`
	Title   Title     `xml:"title,omitempty"`
	Section []Section `xml:"section"`
}

// Title represents a title element
type Title struct {
	Paragraph []Paragraph `xml:"p"`
}

// Section represents a section of the book
type Section struct {
	Title     *Title      `xml:"title,omitempty"`
	Section   []Section   `xml:"section"`
	Paragraph []Paragraph `xml:"p"`
	Poem      []Poem      `xml:"poem,omitempty"`
	Cite      []Cite      `xml:"cite,omitempty"`
	EmptyLine []EmptyLine `xml:"empty-line"`
}

// Paragraph represents a paragraph
type Paragraph struct {
	Text     string     `xml:",chardata"`
	Strong   []Strong   `xml:"strong"`
	Emphasis []Emphasis `xml:"emphasis"`
	Image    []Image    `xml:"image,omitempty"`
	Link     []Link     `xml:"a,omitempty"`
}

// Strong represents bold text (can contain nested elements)
type Strong struct {
	Text     string     `xml:",chardata"`
	Strong   []Strong   `xml:"strong,omitempty"`
	Emphasis []Emphasis `xml:"emphasis,omitempty"`
	Link     []Link     `xml:"a,omitempty"`
}

// Emphasis represents italic text (can contain nested elements)
type Emphasis struct {
	Text     string     `xml:",chardata"`
	Strong   []Strong   `xml:"strong,omitempty"`
	Emphasis []Emphasis `xml:"emphasis,omitempty"`
	Link     []Link     `xml:"a,omitempty"`
}

// Image represents an image reference
type Image struct {
	Href string `xml:"http://www.w3.org/1999/xlink href,attr"`
}

// Link represents a hyperlink
type Link struct {
	Href string `xml:"http://www.w3.org/1999/xlink href,attr"`
	Type string `xml:"type,attr,omitempty"`
	Text string `xml:",chardata"`
}

// Poem represents a poem
type Poem struct {
	Title      *Title   `xml:"title,omitempty"`
	Stanza     []Stanza `xml:"stanza"`
	Date       string   `xml:"date,omitempty"`
	TextAuthor []Author `xml:"text-author,omitempty"`
}

// Stanza represents a stanza in a poem
type Stanza struct {
	Verse []Verse `xml:"v"`
}

// Verse represents a verse line
type Verse struct {
	Text string `xml:",chardata"`
}

// Cite represents a citation
type Cite struct {
	TextAuthor []Author    `xml:"text-author,omitempty"`
	Paragraph  []Paragraph `xml:"p"`
}

// EmptyLine represents an empty line
type EmptyLine struct{}

// Binary represents binary data (images)
type Binary struct {
	ID          string `xml:"id,attr"`
	ContentType string `xml:"content-type,attr"`
	Data        string `xml:",chardata"`
}
