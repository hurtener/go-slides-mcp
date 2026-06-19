// Package markdown parses markdown source into Deckard IR leaf nodes using
// only the Go standard library. It supports the subset the Deckard authoring
// helpers need: ATX headings, bullet and ordered lists, block quotes, and
// plain paragraphs. Inline formatting is emitted as a single TextRun per
// element (no bold / italic parsing in this version).
package markdown

import (
	"bufio"
	"strings"

	"github.com/hurtener/go-slides-mcp/internal/contracts"
)

// Parse turns markdown source into Deckard IR leaf nodes (stdlib only). It
// returns the ordered nodes and any non-fatal warnings. Supported blocks:
//
//	#  /  ##  /  ###  ...                  -> Heading{Level 1..6, Text}
//	- x  /  * x                           -> List{Kind: ListBullet, Items}
//	1. x                                  -> List{Kind: ListNumber, Items}
//	> x                                   -> Quote{Text}
//	blank-separated text                  -> Prose{Paragraphs: [...]}
//
// Consecutive lines of the same block kind fold into ONE node; a blank line
// or a block-type change flushes the current accumulator. Lines that match
// none of the recognisers fold into Prose with a warning.
func Parse(md string) (nodes []contracts.SlideNode, warnings []string) {
	sc := bufio.NewScanner(strings.NewReader(md))
	sc.Buffer(make([]byte, 0, 4096), 1<<20)

	var (
		curKind     blockKind
		bulletItems []contracts.ListItem
		numberItems []contracts.ListItem
		quoteText   []string
		proseLines  []string
	)

	flush := func() {
		switch curKind {
		case blockBullet:
			if len(bulletItems) > 0 {
				nodes = append(nodes, &contracts.List{
					Kind:  contracts.ListBullet,
					Items: bulletItems,
				})
			}
			bulletItems = nil
		case blockNumber:
			if len(numberItems) > 0 {
				nodes = append(nodes, &contracts.List{
					Kind:  contracts.ListNumber,
					Items: numberItems,
				})
			}
			numberItems = nil
		case blockQuote:
			if len(quoteText) > 0 {
				nodes = append(nodes, &contracts.Quote{
					Text: contracts.RichText{{Text: strings.Join(quoteText, " ")}},
				})
			}
			quoteText = nil
		case blockProse:
			if len(proseLines) > 0 {
				paragraphs := splitParagraphs(proseLines)
				paras := make([]contracts.RichText, len(paragraphs))
				for i, p := range paragraphs {
					paras[i] = contracts.RichText{{Text: p}}
				}
				nodes = append(nodes, &contracts.Prose{Paragraphs: paras})
			}
			proseLines = nil
		}
		curKind = blockNone
	}

	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), " \t")
		switch {
		case line == "":
			flush()
		case isATXHeading(line):
			flush()
			level, text := parseHeading(line)
			nodes = append(nodes, &contracts.Heading{
				Level: level,
				Text:  contracts.RichText{{Text: text}},
			})
		case isBulletItem(line):
			if curKind != blockBullet {
				flush()
				curKind = blockBullet
			}
			bulletItems = append(bulletItems, contracts.ListItem{
				Text: contracts.RichText{{Text: strings.TrimSpace(line[1:])}},
			})
		case isNumberItem(line):
			if curKind != blockNumber {
				flush()
				curKind = blockNumber
			}
			text := strings.TrimSpace(line)
			if i := strings.IndexByte(text, '.'); i >= 0 {
				text = strings.TrimSpace(text[i+1:])
			}
			numberItems = append(numberItems, contracts.ListItem{
				Text: contracts.RichText{{Text: text}},
			})
		case isBlockQuote(line):
			if curKind != blockQuote {
				flush()
				curKind = blockQuote
			}
			quoteText = append(quoteText, strings.TrimSpace(line[1:]))
		default:
			if curKind != blockProse {
				flush()
				curKind = blockProse
			}
			proseLines = append(proseLines, line)
		}
	}
	flush()

	return nodes, warnings
}

type blockKind int

const (
	blockNone blockKind = iota
	blockProse
	blockBullet
	blockNumber
	blockQuote
)

func isATXHeading(line string) bool {
	if !strings.HasPrefix(line, "#") {
		return false
	}
	i := 0
	for i < len(line) && line[i] == '#' && i < 6 {
		i++
	}
	if i == 0 || i > 6 {
		return false
	}
	if i == len(line) {
		return true
	}
	return line[i] == ' '
}

func parseHeading(line string) (level int, text string) {
	i := 0
	for i < len(line) && line[i] == '#' {
		i++
	}
	level = i
	text = strings.TrimSpace(line[i:])
	return level, text
}

func isBulletItem(line string) bool {
	if len(line) < 2 {
		return false
	}
	if line[0] != '-' && line[0] != '*' {
		return false
	}
	return line[1] == ' '
}

func isNumberItem(line string) bool {
	i := 0
	for i < len(line) && line[i] >= '0' && line[i] <= '9' {
		i++
	}
	if i == 0 || i+1 >= len(line) || line[i] != '.' || line[i+1] != ' ' {
		return false
	}
	return true
}

func isBlockQuote(line string) bool {
	return len(line) > 0 && line[0] == '>'
}

func splitParagraphs(lines []string) []string {
	out := make([]string, 0, len(lines))
	var cur strings.Builder
	for _, l := range lines {
		if l == "" {
			if cur.Len() > 0 {
				out = append(out, strings.TrimSpace(cur.String()))
				cur.Reset()
			}
			continue
		}
		if cur.Len() == 0 {
			cur.WriteString(l)
		} else {
			cur.WriteByte(' ')
			cur.WriteString(l)
		}
	}
	if cur.Len() > 0 {
		out = append(out, strings.TrimSpace(cur.String()))
	}
	return out
}
