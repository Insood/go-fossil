package main

import "image"

type ArtifactData struct {
	bounds image.Rectangle
	ids    []int32
}

func NewArtifactData(bounds image.Rectangle) *ArtifactData {
	return &ArtifactData{
		bounds: bounds,
		ids:    make([]int32, bounds.Dx()*bounds.Dy()),
	}
}

func (data *ArtifactData) Bounds() image.Rectangle {
	if data == nil {
		return image.Rectangle{}
	}

	return data.bounds
}

func (data *ArtifactData) Clone() *ArtifactData {
	if data == nil {
		return nil
	}

	return &ArtifactData{
		bounds: data.bounds,
		ids:    append([]int32(nil), data.ids...),
	}
}

func (data *ArtifactData) SetID(x, y int, id int32) {
	if data == nil || !image.Pt(x, y).In(data.bounds) {
		return
	}

	index := (y-data.bounds.Min.Y)*data.bounds.Dx() + (x - data.bounds.Min.X)
	data.ids[index] = id
}

func (data *ArtifactData) IDAt(x, y int) int32 {
	if data == nil || !image.Pt(x, y).In(data.bounds) {
		return 0
	}

	index := (y-data.bounds.Min.Y)*data.bounds.Dx() + (x - data.bounds.Min.X)
	return data.ids[index]
}
