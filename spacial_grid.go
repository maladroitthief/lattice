package lattice

import (
	"errors"
	"math"
	"strings"

	"github.com/maladroitthief/caravan"
	"github.com/maladroitthief/mosaic"
)

type (
	SpatialGrid[T comparable] struct {
		sb        strings.Builder
		Nodes     [][]spatialGridNode[T]
		SizeX     int
		SizeY     int
		ChunkSize float64
	}

	spatialGridNode[T comparable] struct {
		x     int
		y     int
		items []T
	}
)

var (
	directions         = [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	ErrMaxDepthReached = errors.New("search max depth has been reached")
)

func NewSpatialGrid[T comparable](x, y int, size float64) *SpatialGrid[T] {
	nodes := make([][]spatialGridNode[T], x)
	for i := range nodes {
		nodes[i] = make([]spatialGridNode[T], y)
	}

	return &SpatialGrid[T]{
		sb:        strings.Builder{},
		SizeX:     x,
		SizeY:     y,
		ChunkSize: size,
		Nodes:     nodes,
	}
}

func (sg *SpatialGrid[T]) Size() int {
	size := 0

	for x := range sg.Nodes {
		for y := range sg.Nodes[x] {
			size += len(sg.Nodes[x][y].items)
		}
	}

	return size
}

func (sg *SpatialGrid[T]) Insert(val T, bounds mosaic.Rectangle) {
	minPoint, maxPoint := bounds.MinPoint(), bounds.MaxPoint()
	xMinIndex, yMinIndex := sg.Location(minPoint.X, minPoint.Y)
	xMaxIndex, yMaxIndex := sg.Location(maxPoint.X, maxPoint.Y)

	for x, xn := xMinIndex, xMaxIndex; x <= xn; x++ {
		for y, yn := yMinIndex, yMaxIndex; y <= yn; y++ {
			sg.Nodes[x][y] = sg.Nodes[x][y].Insert(val)
		}
	}
}

func (sg *SpatialGrid[T]) Update(val T, oldBounds, newBounds mosaic.Rectangle) {
	sg.Delete(val, oldBounds)
	sg.Insert(val, newBounds)
}

func (sg *SpatialGrid[T]) Delete(val T, bounds mosaic.Rectangle) {
	minPoint, maxPoint := bounds.MinPoint(), bounds.MaxPoint()
	xMinIndex, yMinIndex := sg.Location(minPoint.X, minPoint.Y)
	xMaxIndex, yMaxIndex := sg.Location(maxPoint.X, maxPoint.Y)

	for x, xn := xMinIndex, xMaxIndex; x <= xn; x++ {
		for y, yn := yMinIndex, yMaxIndex; y <= yn; y++ {
			sg.Nodes[x][y] = sg.Nodes[x][y].Delete(val)
		}
	}
}

func (sg *SpatialGrid[T]) FindNear(bounds mosaic.Rectangle) []T {
	set := map[T]struct{}{}
	items := []T{}
	minPoint, maxPoint := bounds.MinPoint(), bounds.MaxPoint()
	xMinIndex, yMinIndex := sg.Location(minPoint.X, minPoint.Y)
	xMaxIndex, yMaxIndex := sg.Location(maxPoint.X, maxPoint.Y)

	for x, xn := xMinIndex, xMaxIndex; x <= xn; x++ {
		for y, yn := yMinIndex, yMaxIndex; y <= yn; y++ {
			for _, item := range sg.Nodes[x][y].Items() {
				_, ok := set[item]
				if !ok {
					set[item] = struct{}{}
					items = append(items, item)
				}
			}
		}
	}

	return items
}

func (sg *SpatialGrid[T]) Drop() {
	for i := range sg.Nodes {
		sg.Nodes[i] = make([]spatialGridNode[T], sg.SizeY)
	}
}

func (sg *SpatialGrid[T]) Location(x, y float64) (xIndex, yIndex int) {
	xIndex = int(math.Round(x / sg.ChunkSize))
	yIndex = int(math.Round(y / sg.ChunkSize))

	xIndex = max(xIndex, 0)
	xIndex = min(xIndex, sg.SizeX-1)
	yIndex = max(yIndex, 0)
	yIndex = min(yIndex, sg.SizeY-1)

	return xIndex, yIndex
}

func (sg *SpatialGrid[T]) Node(x, y float64) spatialGridNode[T] {
	xIndex, yIndex := sg.Location(x, y)
	return sg.Nodes[xIndex][yIndex]
}

func (sg *SpatialGrid[T]) GetItemsAtLocation(x, y int) []T {
	return sg.Nodes[x][y].Items()
}

func (sg *SpatialGrid[T]) Search(
	x float64,
	y float64,
	maxDepth int,
	process func([]T) error,
) error {
	start := sg.Node(x, y)

	type index struct{ x, y int }
	visited := map[index]struct{}{}

	queue := caravan.NewQueue[spatialGridNode[T]]()
	queue.Enqueue(start)

	currentDepth := 0
	for queue.Len() > 0 {
		if currentDepth > maxDepth {
			return ErrMaxDepthReached
		}

		nodesAtDepth := queue.Len()
		for i := 0; i < nodesAtDepth; i++ {
			currentNode, err := queue.Dequeue()
			if err != nil {
				return err
			}
			_, ok := visited[index{currentNode.x, currentNode.y}]
			if ok {
				continue
			}
			visited[index{currentNode.x, currentNode.y}] = struct{}{}

			err = process(currentNode.Items())
			if err != nil {
				return err
			}

			edges := sg.Edges(currentNode)
			if len(edges) <= 0 {
				continue
			}

			for _, edge := range edges {
				queue.Enqueue(edge)
			}
		}
		currentDepth++
	}

	return nil
}

func (sg *SpatialGrid[T]) Edges(sgn spatialGridNode[T]) []spatialGridNode[T] {
	edges := []spatialGridNode[T]{}
	for _, direction := range directions {
		nextX := sgn.x + direction[0]
		nextY := sgn.y + direction[1]
		if nextX < 0 || nextX >= sg.SizeX {
			continue
		}
		if nextY < 0 || nextY >= sg.SizeY {
			continue
		}

		edges = append(edges, sg.Nodes[nextX][nextY])
	}

	return edges
}

func NewSpatialGridNode[T comparable](x, y int) spatialGridNode[T] {
	return spatialGridNode[T]{
		items: make([]T, 0),
		x:     x,
		y:     y,
	}
}

func (sgn spatialGridNode[T]) Items() []T {
	items := make([]T, len(sgn.items))

	for i := 0; i < len(sgn.items); i++ {
		items[i] = sgn.items[i]
	}

	return items
}

func (sgn spatialGridNode[T]) Insert(item T) spatialGridNode[T] {
	sgn.items = append(
		sgn.items,
		item,
	)
	return sgn
}

func (sgn spatialGridNode[T]) Delete(item T) spatialGridNode[T] {
	for i := 0; i < len(sgn.items); i++ {
		if sgn.items[i] != item {
			continue
		}
		sgn.items[i] = sgn.items[len(sgn.items)-1]
		sgn.items = sgn.items[:len(sgn.items)-1]
	}

	return sgn
}
