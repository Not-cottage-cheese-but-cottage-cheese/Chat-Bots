package types

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
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
	URL      string
	Keywords []string
}

func (i *Image) GetReader() io.Reader {
	if len(i.ImgBytes) > 0 {
		return bytes.NewBuffer(i.ImgBytes)
	} else {
		response, err := http.Get(i.URL)
		if err != nil {
			log.Println(err)
			return bufio.NewReader(nil)
		}

		return response.Body
	}
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

func (d *Deck) GetUniqKeywordWithImage() (keyword string, numer int) {
	keywords := map[string][]int{}

	for i, image := range d.Images {
		for _, keyword := range image.Keywords {
			if _, ok := keywords[keyword]; !ok {
				keywords[keyword] = []int{}
			}

			keywords[keyword] = append(keywords[keyword], i+1)
		}
	}

	for keyword, images := range keywords {
		if len(images) == 1 {
			return keyword, images[0]
		}
	}

	// если вдруг ключевые слова дублируются, то возьмем любую :)
	imageNumber := rand.Intn(len(d.Images))
	keywordNumber := rand.Intn(len(d.Images[imageNumber].Keywords))

	return d.Images[imageNumber].Keywords[keywordNumber], imageNumber
}
