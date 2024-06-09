package lattice_test

import (
	"fmt"
	"math"
	"math/rand"
	"slices"
	"testing"

	"github.com/maladroitthief/lattice"
	"github.com/maladroitthief/mosaic"
)

const (
	ContainerSize = 1000000
	GridX         = 9
	GridY         = 9
	GridSize      = 32.0
)

type Builder struct {
	layout string
	x      int
	y      int
	size   int
}

func setup_grid(sg *lattice.SpatialGrid[int], b Builder) {
	xPos := func(b Builder, i int) float64 {
		return float64((i%b.x)*b.size) + float64(b.size)/2
	}

	yPos := func(b Builder, i int) float64 {
		return float64((i/b.y)*b.size) + float64(b.size)/2
	}

	for i, block := range b.layout {
		switch block {
		case '0':
		case '1':
			sg.Insert(
				lattice.Item[int]{
					1,
					mosaic.NewRectangle(
						mosaic.Vector{X: xPos(b, i), Y: yPos(b, i)},
						float64(b.size),
						float64(b.size),
					),
					1.0,
				},
			)
		case 'x':
			sg.Insert(
				lattice.Item[int]{
					9,
					mosaic.NewRectangle(
						mosaic.Vector{X: xPos(b, i), Y: yPos(b, i)},
						float64(b.size),
						float64(b.size),
					),
					math.Inf(1),
				},
			)
		default:
		}
	}
}

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
				sg.Insert(lattice.Item[int]{param.item, param.bounds, param.multiplier})
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
	type setup struct {
		builder Builder
	}
	tests := []struct {
		name  string
		setup setup
		wants []float64
	}{
		{
			name: "base case",
			setup: setup{
				builder: Builder{
					x:    9,
					y:    9,
					size: 32,
					layout: "" +
						"100000000" +
						"011111110" +
						"010000010" +
						"010101010" +
						"010101010" +
						"010111010" +
						"010000010" +
						"011111010" +
						"000000000",
				},
			},
			wants: []float64{
				1024.0,
				0.0,
				1024.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](
				tt.setup.builder.x,
				tt.setup.builder.y,
				float64(tt.setup.builder.size),
			)
			setup_grid(sg, tt.setup.builder)

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

			setup_grid(sg, tt.setup.builder)
			got = sg.GetLocationWeight(0, 0)
			want = tt.wants[2]
			if want != got {
				t.Error(fmt.Errorf("spatialGrid.GetLocationWeight() [after restore] want: %+v, got: %+v\n", want, got))
			}
		})
	}
}

func Test_spatial_grid_Search(t *testing.T) {
	type setup struct {
		builder Builder
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
		setup  setup
		params params
		want   want
	}{
		{
			name: "shallow search",
			setup: setup{
				builder: Builder{
					x:    9,
					y:    9,
					size: 32,
					layout: "" +
						"000000000" +
						"011111110" +
						"010000010" +
						"010101010" +
						"010101010" +
						"010111010" +
						"010000010" +
						"011111010" +
						"000000000",
				},
			},
			params: params{
				x:     16,
				y:     16,
				depth: 1,
			},
			want: want{
				items: []int{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sg := lattice.NewSpatialGrid[int](
				tt.setup.builder.x,
				tt.setup.builder.y,
				float64(tt.setup.builder.size),
			)
			setup_grid(sg, tt.setup.builder)
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
	type setup struct {
		builder Builder
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
		setup  setup
		params params
		want   want
	}{
		{
			name: "simple path",
			setup: setup{
				builder: Builder{
					x:    9,
					y:    9,
					size: 32,
					layout: "" +
						"100000000" +
						"011111110" +
						"010000010" +
						"010101010" +
						"010101010" +
						"010111010" +
						"010000010" +
						"011111010" +
						"000000000",
				},
			},
			params: params{
				start: mosaic.NewVector(4, 4),
				end:   mosaic.NewVector(4, 2),
				depth: 32,
			},
			want: want{
				path: []mosaic.Vector{
					{X: 4, Y: 4},
					{X: 4, Y: 3},
					{X: 4, Y: 2},
				},
				err: nil,
			},
		},
		{
			name: "max depth",
			setup: setup{
				builder: Builder{
					x:    9,
					y:    9,
					size: 32,
					layout: "" +
						"000000000" +
						"0xxxxxxx0" +
						"0x00000x0" +
						"0x0x0x0x0" +
						"0x0x0x0x0" +
						"0x0xxx0x0" +
						"0x00000x0" +
						"0xxxxx0x0" +
						"000000000",
				},
			},
			params: params{
				start: mosaic.NewVector(0, 0),
				end:   mosaic.NewVector(4, 4),
				depth: 10,
			},
			want: want{
				path: []mosaic.Vector{},
				err:  lattice.ErrMaxDepthReached,
			},
		},
		{
			name: "hard path",
			setup: setup{
				builder: Builder{
					x:    9,
					y:    9,
					size: 32,
					layout: "" +
						"000000000" +
						"0xxxxxxx0" +
						"0x00000x0" +
						"0x0x0x0x0" +
						"0x0x0x0x0" +
						"0x0xxx0x0" +
						"0x00000x0" +
						"0x0xxx0x0" +
						"000000000",
				},
			},
			params: params{
				start: mosaic.NewVector(0, 0),
				end:   mosaic.NewVector(4, 4),
				depth: 64,
			},
			want: want{
				path: []mosaic.Vector{
					{X: 0, Y: 0},
					{X: 0, Y: 1},
					{X: 0, Y: 2},
					{X: 0, Y: 3},
					{X: 0, Y: 4},
					{X: 0, Y: 5},
					{X: 0, Y: 6},
					{X: 0, Y: 7},
					{X: 0, Y: 8},
					{X: 1, Y: 8},
					{X: 2, Y: 8},
					{X: 2, Y: 7},
					{X: 2, Y: 6},
					{X: 2, Y: 5},
					{X: 2, Y: 4},
					{X: 2, Y: 3},
					{X: 2, Y: 2},
					{X: 3, Y: 2},
					{X: 4, Y: 2},
					{X: 4, Y: 3},
					{X: 4, Y: 4},
				},
				err: nil,
			},
		},
		{
			name: "long path",
			setup: setup{
				builder: Builder{
					x:    18,
					y:    18,
					size: 32,
					layout: "" +
						"0xxxxxx00xxxx0x0x0" +
						"00000000000000x0x0" +
						"0xxxxxx00xxxx0xx00" +
						"00x00000000000x000" +
						"00x0xx00xxxxxxx000" +
						"00x000000000000000" +
						"0xxx00000x0x0x0x00" +
						"00x0000000x0000000" +
						"000xx00000xxx0x000" +
						"00x000x00000x0x0x0" +
						"0x00xxxx00xxxxxxx0" +
						"0x00x00x00x00000x0" +
						"0000x00x00x0x0x0x0" +
						"0x00000000x0x0x0x0" +
						"00x0x00x00x0xxx0x0" +
						"0xx0000x00x00000x0" +
						"0x00x00x00x0xxx0x0" +
						"000000000000000000",
				},
			},
			params: params{
				start: mosaic.NewVector(9, 9),
				end:   mosaic.NewVector(13, 13),
				depth: 32,
			},
			want: want{
				path: []mosaic.Vector{
					{X: 9, Y: 9},
					{X: 9, Y: 10},
					{X: 9, Y: 11},
					{X: 9, Y: 12},
					{X: 9, Y: 13},
					{X: 9, Y: 14},
					{X: 9, Y: 15},
					{X: 9, Y: 16},
					{X: 9, Y: 17},
					{X: 10, Y: 17},
					{X: 11, Y: 17},
					{X: 11, Y: 16},
					{X: 11, Y: 15},
					{X: 11, Y: 14},
					{X: 11, Y: 13},
					{X: 11, Y: 12},
					{X: 11, Y: 11},
					{X: 12, Y: 11},
					{X: 13, Y: 11},
					{X: 13, Y: 12},
					{X: 13, Y: 13},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scale := func(b Builder, v int) float64 {
				return float64(v*b.size) + float64(b.size)/2
			}
			sg := lattice.NewSpatialGrid[int](
				tt.setup.builder.x,
				tt.setup.builder.y,
				float64(tt.setup.builder.size),
			)
			setup_grid(sg, tt.setup.builder)
			start := mosaic.Vector{
				X: scale(tt.setup.builder, int(tt.params.start.X)),
				Y: scale(tt.setup.builder, int(tt.params.start.Y)),
			}
			end := mosaic.Vector{
				X: scale(tt.setup.builder, int(tt.params.end.X)),
				Y: scale(tt.setup.builder, int(tt.params.end.Y)),
			}
			got, err := sg.WeightedSearch(start, end, tt.params.depth)
			if err != tt.want.err {
				t.Error(
					fmt.Errorf(
						"spatialGrid.WeightedSearch() error. want error: %+v, got error: %+v\n",
						tt.want.err,
						err,
					),
				)
			}
			path := make([]mosaic.Vector, len(tt.want.path))
			for i := 0; i < len(path); i++ {
				path[i] = mosaic.Vector{
					X: scale(tt.setup.builder, int(tt.want.path[i].X)),
					Y: scale(tt.setup.builder, int(tt.want.path[i].Y)),
				}
			}

			if !slices.Equal(path, got) {
				t.Error(fmt.Errorf("spatialGrid.WeightedSearch() want: %+v, got: %+v\n", path, got))
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
			lattice.Item[int]{
				rand.Int(),
				mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
				rand.Float64(),
			},
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
			lattice.Item[int]{
				rand.Int(),
				mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
				rand.Float64(),
			},
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
		sg.Insert(lattice.Item[int]{value, bounds, rand.Float64()})
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
			lattice.Item[int]{
				rand.Int(),
				bounds,
				rand.Float64(),
			},
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
			lattice.Item[int]{
				rand.Int(),
				mosaic.NewRectangle(mosaic.Vector{X: x0, Y: y0}, sizeX, sizeY),
				rand.Float64(),
			},
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
