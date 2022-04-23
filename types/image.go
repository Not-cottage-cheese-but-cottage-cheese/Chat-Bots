package types

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Image struct {
	ID       string
	Name     string
	ImgBytes []byte
	Keywords []string
}

type Deck struct {
	Images []*Image
}

func getKeywords(path string) (map[string][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	result := map[string][]string{}

	defer f.Close()

	scanner := bufio.NewScanner(f)
	re := regexp.MustCompile(`\s+`)
	for scanner.Scan() {
		content := re.Split(scanner.Text(), -1)
		result[content[0]] = content[1:]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func NewDeckFromFiles(path string, descPath string) (*Deck, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	keywords, err := getKeywords(descPath)
	if err != nil {
		return nil, err
	}

	deck := &Deck{}
	for _, file := range files {
		fileBytes, err := ioutil.ReadFile(
			filepath.Join(path, file.Name()),
		)
		if err != nil {
			return nil, err
		}

		deck.Images = append(deck.Images, &Image{
			ID:       uuid.NewString(),
			Name:     strings.TrimSuffix(file.Name(), filepath.Ext(file.Name())),
			ImgBytes: fileBytes,
			Keywords: keywords[file.Name()],
		})
	}

	return deck, nil
}

func (d *Deck) GetCards(count int) *Deck {
	if count > len(d.Images) {
		count = len(d.Images)
	}

	result := &Deck{
		Images: d.Images[:count],
	}
	d.Images = d.Images[count:]

	return result
}

func (d *Deck) GetUniqKeywordWithImage() (keyword string, image *Image) {
	keywords := map[string][]*Image{}

	for _, image := range d.Images {
		for _, keyword := range image.Keywords {
			if _, ok := keywords[keyword]; !ok {
				keywords[keyword] = []*Image{}
			}

			keywords[keyword] = append(keywords[keyword], image)
		}
	}

	for keyword, images := range keywords {
		if len(images) == 1 {
			return keyword, images[0]
		}
	}

	return "", nil
}
