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
		item       int
		bounds     mosaic.Rectangle
		multiplier float64
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
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2), multiplier: 1.0},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2), multiplier: 1.0},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2), multiplier: 1.0},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2), multiplier: 1.0},
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
				sg.Insert(param.item, param.bounds, param.multiplier)
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
	type item struct {
		value      int
		bounds     mosaic.Rectangle
		multiplier float64
	}
	type fields struct {
		x     int
		y     int
		size  float64
		items []item
	}
	type params struct {
		x     float64
		y     float64
		depth int
	}
	type want struct {
		items []int
	}
	tests := []struct {
		name   string
		fields fields
		params params
		want   want
	}{
		{
			name: "shallow search",
			fields: fields{
				x:    4,
				y:    4,
				size: 8,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2), multiplier: 1.0},
					{value: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2), multiplier: 1.0},
					{value: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2), multiplier: 1.0},
					{value: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2), multiplier: 1.0},
				},
			},
			params: params{
				x:     20,
				y:     20,
				depth: 1,
			},
			want: want{
				items: []int{3},
			},
		},
		{
			name: "wide search",
			fields: fields{
				x:    32,
				y:    32,
				size: 8,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2), multiplier: 1.0},
					{value: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2), multiplier: 1.0},
					{value: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2), multiplier: 1.0},
					{value: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2), multiplier: 1.0},
				},
			},
			params: params{
				x:     20,
				y:     20,
				depth: 5,
			},
			want: want{
				items: []int{1, 2, 3, 4},
			},
		},
		{
			name: "medium search",
			fields: fields{
				x:    32,
				y:    32,
				size: 8,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2), multiplier: 1.0},
					{value: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2), multiplier: 1.0},
					{value: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2), multiplier: 1.0},
					{value: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2), multiplier: 1.0},
				},
			},
			params: params{
				x:     0,
				y:     0,
				depth: 2,
			},
			want: want{
				items: []int{1, 2},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, item := range tt.fields.items {
				sg.Insert(item.value, item.bounds, item.multiplier)
			}

			got := []int{}
			want := tt.want.items
			check := func(items []int) error {
				got = append(got, items...)
				return nil
			}

			sg.Search(tt.params.x, tt.params.y, tt.params.depth, check)
			slices.Sort(want)
			slices.Sort(got)

			if !slices.Equal(want, got) {
				t.Error(fmt.Errorf("spatialGrid.Search() want: %+v, got: %+v\n", want, got))
			}
		})
	}
}

func Test_spatial_grid_WeightedSearch(t *testing.T) {
	type item struct {
		value      int
		bounds     mosaic.Rectangle
		multiplier float64
	}
	type fields struct {
		x     int
		y     int
		size  float64
		items []item
	}
	type params struct {
		start mosaic.Vector
		end   mosaic.Vector
	}
	type want struct {
		path []mosaic.Vector
		err  error
	}
	tests := []struct {
		name   string
		fields fields
		params params
		want   want
	}{
		{
			name: "simple path",
			fields: fields{
				x:     4,
				y:     4,
				size:  10,
				items: []item{},
			},
			params: params{
				start: mosaic.NewVector(5, 5),
				end:   mosaic.NewVector(10, 5),
			},
			want: want{
				path: []mosaic.Vector{
					{X: 5, Y: 5},
					{X: 15, Y: 5},
				},
				err: nil,
			},
		},
		{
			name: "long path",
			fields: fields{
				x:     4,
				y:     4,
				size:  10,
				items: []item{},
			},
			params: params{
				start: mosaic.NewVector(5, 5),
				end:   mosaic.NewVector(35, 5),
			},
			want: want{
				path: []mosaic.Vector{
					{X: 5, Y: 5},
					{X: 15, Y: 5},
					{X: 25, Y: 5},
					{X: 35, Y: 5},
				},
				err: nil,
			},
		},
		{
			name: "long path with obstacles",
			fields: fields{
				x:    4,
				y:    4,
				size: 10,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 15, Y: 5}, 8, 8), multiplier: 1.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 25, Y: 5}, 8, 8), multiplier: 1.0},
				},
			},
			params: params{
				start: mosaic.NewVector(5, 5),
				end:   mosaic.NewVector(35, 5),
			},
			want: want{
				path: []mosaic.Vector{
					{X: 5, Y: 5},
					{X: 5, Y: 15},
					{X: 15, Y: 15},
					{X: 25, Y: 15},
					{X: 35, Y: 15},
					{X: 35, Y: 5},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, item := range tt.fields.items {
				sg.Insert(item.value, item.bounds, item.multiplier)
			}

			got, err := sg.WeightedSearch(tt.params.start, tt.params.end)

			if err != tt.want.err {
				t.Error(
					fmt.Errorf(
						"spatialGrid.WeightedSearch() error. want error: %+v, got error: %+v\n",
						tt.want.err,
						err,
					),
				)
			}

			if !slices.Equal(tt.want.path, got) {
				t.Error(fmt.Errorf("spatialGrid.WeightedSearch() want: %+v, got: %+v\n", tt.want.path, got))
			}
		})
	}
}
