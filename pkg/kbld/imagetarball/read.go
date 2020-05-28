package tarball

import (
	"io"
	"io/ioutil"

	regv1 "github.com/google/go-containerregistry/pkg/v1"
)

type ImageWithRef interface {
	regv1.Image
	Ref() string
}

type ImageIndexWithRef interface {
	regv1.ImageIndex
	Ref() string
}

type LayerContents interface {
	Open() (io.ReadCloser, error)
}

type LayerProvider interface {
	FindLayer(ImageLayerTarDescriptor) (LayerContents, error)
}

type TarImageOrIndex struct {
	Image *ImageWithRef
	Index *ImageIndexWithRef
}

func (t TarImageOrIndex) Digest() (regv1.Hash, error) {
	switch {
	case t.Image != nil:
		return (*t.Image).Digest()
	case t.Index != nil:
		return (*t.Index).Digest()
	default:
		panic("Unknown item")
	}
}

func (t TarImageOrIndex) Ref() string {
	switch {
	case t.Image != nil:
		return (*t.Image).Ref()
	case t.Index != nil:
		return (*t.Index).Ref()
	default:
		panic("Unknown item")
	}
}

func MultiRefReadFromFile(path string) ([]TarImageOrIndex, error) {
	file := tarFile{path}

	manifestFile, err := file.Chunk("manifest.json").Open()
	if err != nil {
		return nil, err
	}

	manifestBytes, err := ioutil.ReadAll(manifestFile)
	if err != nil {
		return nil, err
	}

	tds, err := NewTarDescriptorsFromBytes(manifestBytes)
	if err != nil {
		return nil, err
	}

	return ReadFromTds(tds, file), nil
}

func ReadFromTds(tds *TarDescriptors, layerProvider LayerProvider) []TarImageOrIndex {
	var result []TarImageOrIndex

	for _, td := range tds.tds {
		switch {
		case td.Image != nil:
			var img ImageWithRef = tarImage{*td.Image, layerProvider}
			result = append(result, TarImageOrIndex{Image: &img})

		case td.ImageIndex != nil:
			idx := buildIndex(*td.ImageIndex, layerProvider)
			result = append(result, TarImageOrIndex{Index: &idx})

		default:
			panic("Unknown item")
		}
	}

	return result
}

func buildIndex(iitd ImageIndexTarDescriptor, layerProvider LayerProvider) ImageIndexWithRef {
	var images []regv1.Image
	var indexes []regv1.ImageIndex

	for _, imgTD := range iitd.Images {
		images = append(images, tarImage{imgTD, layerProvider})
	}
	for _, indexTD := range iitd.Indexes {
		indexes = append(indexes, buildIndex(indexTD, layerProvider))
	}

	return tarImageIndex{iitd, images, indexes}
}
