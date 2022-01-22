package block

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// Range describes an area of code, including the filename it is present in and the lin numbers the code occupies
type Range struct {
	Filename  string `json:"filename"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

// String creates a human-readable summary of the range
func (r Range) String() string {
	if r.StartLine != r.EndLine {
		return fmt.Sprintf("%s:%d-%d", r.Filename, r.StartLine, r.EndLine)
	}
	return fmt.Sprintf("%s:%d", r.Filename, r.StartLine)
}

func (r Range) ReadLines(includeCommentsAfterLines bool) (lines []string, comments []string, err error) {
	data, err := ioutil.ReadFile(r.Filename)
	if err != nil {
		return nil, nil, err
	}
	rawLines := strings.Split(string(data), "\n")

	allLines := []string{""}
	for _, rawLine := range rawLines {
		allLines = append(allLines, strings.Trim(rawLine, "\r"))
	}

	var inComment bool

	for commentStart := r.StartLine - 1; commentStart > 0; commentStart-- {
		line := strings.TrimSpace(allLines[commentStart])
		if strings.HasPrefix(line, "/*") {
			line = line[2:]
			inComment = true
		} else if strings.HasPrefix(line, "#") {
			line = line[1:]
		} else if strings.HasPrefix(line, "//") {
			line = line[2:]
		} else if !inComment {
			break
		}

		if strings.HasSuffix(line, "*/") {
			inComment = false
			line = line[:strings.LastIndex(line, "*/")]
		}

		comments = append([]string{line}, comments...)
	}
	if includeCommentsAfterLines {
		comments = append(comments, r.readInlineComments(allLines)...)
	}

	for i := r.StartLine; i < r.EndLine; i++ {
		lines = append(lines, allLines[i])
	}

	return lines, comments, nil
}

func (r Range) readInlineComments(allLines []string) []string {
	var comments []string
	for commentStart := r.StartLine; commentStart <= r.EndLine; commentStart++ {
		line := strings.TrimSpace(allLines[commentStart])
		if strings.Contains(line, "#") {
			comments = append(comments, line[strings.Index(line, "#")+1:])
		} else if strings.Contains(line, "//") {
			comments = append(comments, line[strings.Index(line, "//")+2:])
		} else if strings.Contains(line, "/*") {
			line = line[strings.Index(line, "/*")+2:]
			if strings.Contains(line, "*/") {
				line = line[:strings.LastIndex(line, "*/")]
			}
			comments = append(comments, line)
		}
	}
	return comments
}
