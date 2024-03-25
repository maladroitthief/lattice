package lattice_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/maladroitthief/lattice"
	"github.com/maladroitthief/mosaic"
)

func Test_spatial_grid_Insert(t *testing.T) {
	type fields struct {
		x    int
		y    int
		size float64
	}
	type params struct {
		item   int
		bounds mosaic.Rectangle
	}
	type wants struct {
		x    int
		y    int
		item int
	}
	tests := []struct {
		name   string
		fields fields
		params []params
		wants  []wants
	}{
		{
			name:   "single insert",
			fields: fields{x: 4, y: 4, size: 8},
			params: []params{
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2)},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2)},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2)},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2)},
			},
			wants: []wants{
				{x: 0, y: 0, item: 1},
				{x: 1, y: 1, item: 2},
				{x: 2, y: 2, item: 3},
				{x: 3, y: 3, item: 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, param := range tt.params {
				sg.Insert(param.item, param.bounds)
			}

			for _, want := range tt.wants {
				got := slices.Contains(sg.Nodes[want.x][want.y].Items(), want.item)
				if !got {
					t.Errorf("spatialGrid.Insert() did not find what we want: %+v", want)
				}
			}
		})
	}
}

func Test_spatial_grid_Search(t *testing.T) {
	type fields struct {
		x    int
		y    int
		size float64
	}
	type params struct {
		item   int
		bounds mosaic.Rectangle
	}
	type wants struct {
		x     float64
		y     float64
		depth int
		items []int
	}
	tests := []struct {
		name   string
		fields fields
		params []params
		wants  []wants
	}{
		{
			name:   "shallow search",
			fields: fields{x: 4, y: 4, size: 8},
			params: []params{
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2)},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2)},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2)},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2)},
			},
			wants: []wants{
				{x: 20, y: 20, depth: 1, items: []int{3}},
			},
		},
		{
			name:   "wide search",
			fields: fields{x: 32, y: 32, size: 8},
			params: []params{
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2)},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2)},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2)},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2)},
			},
			wants: []wants{
				{x: 20, y: 20, depth: 5, items: []int{1, 2, 3, 4}},
			},
		},
		{
			name:   "wide search",
			fields: fields{x: 32, y: 32, size: 8},
			params: []params{
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2)},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2)},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2)},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2)},
			},
			wants: []wants{
				{x: 0, y: 0, depth: 2, items: []int{1, 2}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, param := range tt.params {
				sg.Insert(param.item, param.bounds)
			}

			for _, want := range tt.wants {
				got := []int{}
				check := func(items []int) error {
					got = append(got, items...)
					return nil
				}

				sg.Search(want.x, want.y, want.depth, check)
				slices.Sort(want.items)
				slices.Sort(got)

				if !slices.Equal(want.items, got) {
					t.Error(fmt.Errorf("spatialGrid.Search() want: %+v, got: %+v\n", want.items, got))
				}
			}
		})
	}
}
