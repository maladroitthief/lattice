package lattice_test

import (
	"fmt"
	"math/rand"
	"slices"
	"testing"

	"github.com/maladroitthief/lattice"
	"github.com/maladroitthief/mosaic"
)

const (
	ContainerSize = 1000000
	GridX         = 8
	GridY         = 8
	GridSize      = 32.0
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
				got := slices.Contains(sg.Nodes[want.x][want.y].Values(), want.item)
				if !got {
					t.Errorf("spatialGrid.Insert() did not find what we want: %+v", want)
				}
			}
		})
	}
}

func Test_spatial_grid_GetLocationWeight(t *testing.T) {
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
	tests := []struct {
		name   string
		fields fields
		params []params
		wants  []float64
	}{
		{
			name:   "base case",
			fields: fields{x: 4, y: 4, size: 8},
			params: []params{
				{item: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 4, Y: 4}, 2, 2), multiplier: 1.0},
				{item: 2, bounds: mosaic.NewRectangle(mosaic.Vector{X: 12, Y: 12}, 2, 2), multiplier: 1.0},
				{item: 3, bounds: mosaic.NewRectangle(mosaic.Vector{X: 20, Y: 20}, 2, 2), multiplier: 1.0},
				{item: 4, bounds: mosaic.NewRectangle(mosaic.Vector{X: 28, Y: 28}, 2, 2), multiplier: 1.0},
			},
			wants: []float64{
				4.0,
				0.0,
				4.0,
				0.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, param := range tt.params {
				sg.Insert(param.item, param.bounds, param.multiplier)
			}

			got := sg.GetLocationWeight(0, 0)
			want := tt.wants[0]
			if want != got {
				t.Error(fmt.Errorf("spatialGrid.GetLocationWeight() want: %+v, got: %+v\n", want, got))
			}

			sg.Drop()
			got = sg.GetLocationWeight(0, 0)
			want = tt.wants[1]
			if want != got {
				t.Error(fmt.Errorf("spatialGrid.GetLocationWeight() [after drop] want: %+v, got: %+v\n", want, got))
			}

			for _, param := range tt.params {
				sg.Insert(param.item, param.bounds, param.multiplier)
			}
			got = sg.GetLocationWeight(0, 0)
			want = tt.wants[2]
			if want != got {
				t.Error(fmt.Errorf("spatialGrid.GetLocationWeight() [after restore] want: %+v, got: %+v\n", want, got))
			}

			sg.Delete(tt.params[0].item, tt.params[0].bounds)
			got = sg.GetLocationWeight(0, 0)
			want = tt.wants[3]
			if want != got {
				t.Error(fmt.Errorf("spatialGrid.GetLocationWeight() [after delete] want: %+v, got: %+v\n", want, got))
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
		{
			name: "long path with weighted obstacles",
			fields: fields{
				x:    8,
				y:    8,
				size: 32,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 0, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 0, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 0, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 0, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 112}, 32, 32), multiplier: 10.0},
				},
			},
			params: params{
				x:     128,
				y:     128,
				depth: 1,
			},
			want: want{
				items: []int{1, 1, 1, 1, 1},
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
		depth int
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
				depth: 10,
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
				depth: 10,
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
				depth: 10,
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
		{
			name: "long path with weighted obstacles",
			fields: fields{
				x:    8,
				y:    8,
				size: 32,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 80, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 208}, 32, 32), multiplier: 10.0},
				},
			},
			params: params{
				start: mosaic.NewVector(48, 48),
				end:   mosaic.NewVector(208, 208),
				depth: 64,
			},
			want: want{
				path: []mosaic.Vector{
					{X: 48, Y: 48},
					{X: 48, Y: 16},
					{X: 80, Y: 16},
					{X: 112, Y: 16},
					{X: 144, Y: 16},
					{X: 144, Y: 48},
					{X: 144, Y: 80},
					{X: 144, Y: 112},
					{X: 176, Y: 112},
					{X: 176, Y: 144},
					{X: 176, Y: 176},
					{X: 176, Y: 208},
					{X: 208, Y: 208},
				},
				err: nil,
			},
		},
		{
			name: "long path with low depth",
			fields: fields{
				x:    8,
				y:    8,
				size: 32,
				items: []item{
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 48, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 80, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 48}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 176, Y: 80}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 112, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 112}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 144}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 208, Y: 176}, 32, 32), multiplier: 10.0},
					{value: 1, bounds: mosaic.NewRectangle(mosaic.Vector{X: 144, Y: 208}, 32, 32), multiplier: 10.0},
				},
			},
			params: params{
				start: mosaic.NewVector(48, 48),
				end:   mosaic.NewVector(208, 208),
				depth: 5,
			},
			want: want{
				path: []mosaic.Vector{},
				err:  lattice.ErrMaxDepthReached,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](tt.fields.x, tt.fields.y, tt.fields.size)
			for _, item := range tt.fields.items {
				sg.Insert(item.value, item.bounds, item.multiplier)
			}
			sg.Drop()
			for _, item := range tt.fields.items {
				sg.Insert(item.value, item.bounds, item.multiplier)
			}

			got, err := sg.WeightedSearch(tt.params.start, tt.params.end, tt.params.depth)

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

func BenchmarkSpatialGridSize(b *testing.B) {
	sg := lattice.NewSpatialGrid[int](GridX, GridY, GridSize)
	for i := 0; i < ContainerSize; i++ {
		x0 := float64(rand.Intn(GridX) * rand.Intn(int(GridSize)))
		y0 := float64(rand.Intn(GridY) * rand.Intn(int(GridSize)))
		sizeX := GridSize * rand.Float64()
		sizeY := GridSize * rand.Float64()
		sg.Insert(
			rand.Int(),
			mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
			rand.Float64(),
		)
	}

	for n := 0; n < b.N; n++ {
		sg.Size()
	}
}

func BenchmarkSpatialGridInsert(b *testing.B) {
	sg := lattice.NewSpatialGrid[int](GridX, GridY, GridSize)
	for n := 0; n < b.N; n++ {
		x0 := float64(rand.Intn(GridX) * rand.Intn(int(GridSize)))
		y0 := float64(rand.Intn(GridY) * rand.Intn(int(GridSize)))
		sizeX := GridSize * rand.Float64()
		sizeY := GridSize * rand.Float64()
		sg.Insert(
			rand.Int(),
			mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
			rand.Float64(),
		)
	}
}

func BenchmarkSpatialGridDelete(b *testing.B) {
	type entity struct {
		value  int
		bounds mosaic.Rectangle
	}

	sg := lattice.NewSpatialGrid[int](GridX, GridY, GridSize)
	entities := []entity{}
	for i := 0; i < ContainerSize; i++ {
		x0 := float64(rand.Intn(GridX) * rand.Intn(int(GridSize)))
		y0 := float64(rand.Intn(GridY) * rand.Intn(int(GridSize)))
		sizeX := GridSize * rand.Float64()
		sizeY := GridSize * rand.Float64()

		value := rand.Int()
		bounds := mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY)
		sg.Insert(value, bounds, rand.Float64())
		entities = append(entities, entity{value: value, bounds: bounds})
	}

	for n := 0; n < b.N; n++ {
		if n < ContainerSize {
			index := rand.Intn(len(entities))
			entity := entities[index]

			sg.Delete(entity.value, entity.bounds)

			entities[index] = entities[len(entities)-1]
			entities = entities[:len(entities)-1]
		}
	}
}

func BenchmarkSpatialGridFindNear(b *testing.B) {
	sg := lattice.NewSpatialGrid[int](GridX, GridY, GridSize)
	entities := []mosaic.Rectangle{}

	for i := 0; i < ContainerSize; i++ {
		x0 := float64(rand.Intn(GridX) * rand.Intn(int(GridSize)))
		y0 := float64(rand.Intn(GridY) * rand.Intn(int(GridSize)))
		sizeX := GridSize * rand.Float64()
		sizeY := GridSize * rand.Float64()
		bounds := mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY)
		entities = append(entities, bounds)

		sg.Insert(
			rand.Int(),
			bounds,
			rand.Float64(),
		)
	}

	for n := 0; n < b.N; n++ {
		sg.FindNear(entities[n%len(entities)])
	}
}

func BenchmarkSpatialGridWeightedSearch(b *testing.B) {
	sg := lattice.NewSpatialGrid[int](GridX, GridY, GridSize)
	entities := []mosaic.Vector{}
	for i := 0; i < ContainerSize; i++ {
		x0 := float64(rand.Intn(GridX) * rand.Intn(int(GridSize)))
		y0 := float64(rand.Intn(GridY) * rand.Intn(int(GridSize)))
		sizeX := GridSize * rand.Float64()
		sizeY := GridSize * rand.Float64()
		entities = append(
			entities,
			mosaic.NewVector(
				float64(rand.Intn(GridX)*rand.Intn(int(GridSize))),
				float64(rand.Intn(GridY)*rand.Intn(int(GridSize))),
			),
		)

		sg.Insert(
			rand.Int(),
			mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
			rand.Float64(),
		)
	}

	for n := 0; n < b.N; n++ {
		sg.WeightedSearch(
			entities[n%len(entities)],
			entities[n%len(entities)],
			10,
		)
	}
}
