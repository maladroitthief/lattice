package lattice

import (
	"errors"
	"math"
	"sync"

	"github.com/maladroitthief/caravan"
	"github.com/maladroitthief/mosaic"
)

type (
	SpatialGrid[T comparable] struct {
		Nodes     [][]spatialGridNode[T]
		nodesMu   sync.RWMutex
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
	for iX := range nodes {
		nodes[iX] = make([]spatialGridNode[T], y)
		for iY := range y {
			nodes[iX][iY].x = iX
			nodes[iX][iY].y = iY
		}
	}

	return &SpatialGrid[T]{
		SizeX:     x,
		SizeY:     y,
		ChunkSize: size,
		Nodes:     nodes,
	}
}

func (sg *SpatialGrid[T]) Size() int {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	size := 0
	for x := range sg.Nodes {
		for y := range sg.Nodes[x] {
			size += len(sg.Nodes[x][y].items)
		}
	}

	return size
}

func (sg *SpatialGrid[T]) Insert(val T, bounds mosaic.Rectangle) {
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

	minPoint, maxPoint := bounds.MinPoint(), bounds.MaxPoint()
	xMinIndex, yMinIndex := sg.Location(minPoint.X, minPoint.Y)
	xMaxIndex, yMaxIndex := sg.Location(maxPoint.X, maxPoint.Y)

	for xMin, xMax := xMinIndex, xMaxIndex; xMin <= xMax; xMin++ {
		for yMin, yMax := yMinIndex, yMaxIndex; yMin <= yMax; yMin++ {
			sg.Nodes[xMin][yMin] = sg.Nodes[xMin][yMin].Insert(val)
		}
	}
}

func (sg *SpatialGrid[T]) Update(val T, oldBounds, newBounds mosaic.Rectangle) {
	sg.Delete(val, oldBounds)
	sg.Insert(val, newBounds)
}

func (sg *SpatialGrid[T]) Delete(val T, bounds mosaic.Rectangle) {
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

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
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

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
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

	for iX := range sg.Nodes {
		sg.Nodes[iX] = make([]spatialGridNode[T], sg.SizeY)
		for iY := range sg.SizeY {
			sg.Nodes[iX][iY].x = iX
			sg.Nodes[iX][iY].y = iY
		}
	}
}

func (sg *SpatialGrid[T]) Location(x, y float64) (xIndex, yIndex int) {
	xIndex = int(math.Floor(x / sg.ChunkSize))
	yIndex = int(math.Floor(y / sg.ChunkSize))

	xIndex = max(xIndex, 0)
	xIndex = min(xIndex, sg.SizeX-1)
	yIndex = max(yIndex, 0)
	yIndex = min(yIndex, sg.SizeY-1)

	return xIndex, yIndex
}

func (sg *SpatialGrid[T]) GetItemsAtLocation(x, y int) []T {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	return sg.Nodes[x][y].Items()
}

func (sg *SpatialGrid[T]) Node(x, y float64) spatialGridNode[T] {
	xIndex, yIndex := sg.Location(x, y)
	return sg.Nodes[xIndex][yIndex]
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

func (sg *SpatialGrid[T]) Search(
	x float64,
	y float64,
	maxDepth int,
	process func([]T) error,
) error {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

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

func (sgn *SpatialGrid[T]) WalkGrid(v, w mosaic.Vector) []mosaic.Vector {
	delta := w.Subtract(v)
	nX, nY := math.Abs(delta.X), math.Abs(delta.Y)
	signX, signY := 1.0, 1.0
	if delta.X <= 0 {
		signX = -1
	}
	if delta.Y <= 0 {
		signY = -1
	}
	vector := v.Clone()
	vectors := []mosaic.Vector{vector.Clone()}

	i, j := 0.0, 0.0
	for i < nX || j < nY {
		if (1+2*i)*nY < (1+2*j)*nX {
			vector.X += signX
			i++
		} else {
			vector.Y += signY
			j++
		}
		vectors = append(vectors, vector.Clone())
	}

	return vectors
}
