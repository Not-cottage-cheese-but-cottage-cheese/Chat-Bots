package types

import (
	"io/ioutil"
	"math/rand"
	"path/filepath"

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

func NewDeckFromFiles(path string, descPath string) (*Deck, error) {
	files, err := ioutil.ReadDir(path)
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
			Name:     file.Name(),
			ImgBytes: fileBytes,
			Keywords: []string{},
		})
	}
	return deck, nil
}

func (d *Deck) GetCards(count int) *Deck {
	rand.Shuffle(len(d.Images), func(i, j int) { d.Images[i], d.Images[j] = d.Images[j], d.Images[i] })

	if count > len(d.Images) {
		count = len(d.Images)
	}

	result := &Deck{
		Images: d.Images[:count],
	}
	d.Images = d.Images[count:]

	return result
}
