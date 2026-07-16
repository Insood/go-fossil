package main

import "image"

type ArtifactData struct {
	bounds image.Rectangle
	ids    []uint32
}

func NewArtifactData(bounds image.Rectangle) *ArtifactData {
	return &ArtifactData{
		bounds: bounds,
		ids:    make([]uint32, bounds.Dx()*bounds.Dy()),
	}
}

func (data *ArtifactData) Bounds() image.Rectangle {
	if data == nil {
		return image.Rectangle{}
	}

	return data.bounds
}

func (data *ArtifactData) SetID(x, y int, id uint32) {
	if data == nil || !image.Pt(x, y).In(data.bounds) {
		return
	}

	index := (y-data.bounds.Min.Y)*data.bounds.Dx() + (x - data.bounds.Min.X)
	data.ids[index] = id
}

func (data *ArtifactData) IDAt(x, y int) uint32 {
	if data == nil || !image.Pt(x, y).In(data.bounds) {
		return 0
	}

	index := (y-data.bounds.Min.Y)*data.bounds.Dx() + (x - data.bounds.Min.X)
	return data.ids[index]
}
