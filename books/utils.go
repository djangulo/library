package books

import (
	"regexp"
	"strconv"
	"strings"
)

// Slugify returns a slug-compatible version, separated by slugChar
func Slugify(str string, slugChar string) string {
	body := regexp.MustCompile(`[._\\/!?#$%, ]+`)
	bodyRemove := regexp.MustCompile(`\'+`)
	edge1 := regexp.MustCompile(`^-*`)
	edge2 := regexp.MustCompile(`-*$`)
	multiple := regexp.MustCompile(`[-]{2,}`)
	str = strings.ToLower(str)
	result := body.ReplaceAll([]byte(str), []byte(slugChar))
	result = bodyRemove.ReplaceAll(result, []byte(""))
	result = edge1.ReplaceAll(result, []byte(""))
	result = edge2.ReplaceAll(result, []byte(""))
	result = multiple.ReplaceAll(result, []byte(slugChar))
	return string(result)
}

// GutenbergMeta extract metadata from the gutenberg format
func GutenbergMeta(line string) Book {
	var book Book
	re := regexp.MustCompile(`^\[([a-zA-Z\s'-]+)\,? (by|By)? ([\s.a-zA-Z]+)?\s?([\d]+)?\]$`)
	loc := re.FindAllSubmatch([]byte(line), -1)
	book.Source = NewNullString("gutenberg")

	if len(loc) > 0 {
		if title := string(loc[0][1]); title != "" {
			book.Title = strings.Trim(title, " ")
			book.Slug = Slugify(book.Title, "-")
		}
		if author := string(loc[0][3]); author != "" {
			book.Author = NewNullString(strings.Trim(author, " "))
		} else {
			book.Author = NewNullString("")
		}
		if pubYear := string(loc[0][4]); pubYear != "" {
			year, err := strconv.Atoi(strings.Trim(pubYear, " "))
			if err != nil {
				book.PublicationYear = NewNullInt64(0)
			} else {
				book.PublicationYear = NewNullInt64(int64(year))
			}
		}
	} else {
		edges := regexp.MustCompile(`[\]\[]+`)
		title := edges.ReplaceAllString(line, "")
		book.Title = title
		book.Slug = Slugify(title, "-")
		book.PublicationYear = NewNullInt64(0)
		book.Author = NewNullString("")
	}

	return book
}
