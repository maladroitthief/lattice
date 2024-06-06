package lattice

import (
	"errors"
	"math"
	"sync"

	"github.com/maladroitthief/caravan"
	"github.com/maladroitthief/mosaic"
	"golang.org/x/exp/maps"
)

type (
	SpatialGrid[T comparable] struct {
		Nodes     [][]spatialGridNode[T]
		nodesMu   sync.RWMutex
		SizeX     int
		SizeY     int
		ChunkSize float64
		itemCount int
	}

	spatialGridNode[T comparable] struct {
		x      int
		y      int
		bounds mosaic.Rectangle
		weight float64
		Items  []spatialGridNodeItem[T]
	}

	spatialGridNodeItem[T comparable] struct {
		value      T
		bounds     mosaic.Rectangle
		multiplier float64
		weight     float64
	}
)

var (
	directions         = [][]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}}
	ErrMaxDepthReached = errors.New("search max depth has been reached")
	ErrPathNotFound    = errors.New("weighted search could not find a path")
)

func NewSpatialGrid[T comparable](x, y int, size float64) *SpatialGrid[T] {
	nodes := make([][]spatialGridNode[T], x)
	for iX := range nodes {
		nodes[iX] = make([]spatialGridNode[T], y)
		for iY := range y {
			nodes[iX][iY] = newSpatialGridNode[T](
				iX,
				iY,
				mosaic.NewRectangle(
					mosaic.NewVector(
						(float64(iX)*size)+size/2,
						(float64(iY)*size)+size/2,
					),
					size,
					size,
				),
			)
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
	return sg.itemCount
}

func (sg *SpatialGrid[T]) Insert(val T, bounds mosaic.Rectangle, multiplier float64) {
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

	x, y := sg.Location(bounds.Position.X, bounds.Position.Y)
	sg.Nodes[x][y] = sg.Nodes[x][y].Insert(val, bounds, multiplier)
	sg.itemCount++
}

func (sg *SpatialGrid[T]) Update(val T, oldBounds, newBounds mosaic.Rectangle, multiplier float64) {
	sg.Delete(val, oldBounds)
	sg.Insert(val, newBounds, multiplier)
}

func (sg *SpatialGrid[T]) Delete(val T, bounds mosaic.Rectangle) {
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

	x, y := sg.Location(bounds.Position.X, bounds.Position.Y)
	sg.Nodes[x][y] = sg.Nodes[x][y].Delete(val)

	sg.itemCount--
}

func (sg *SpatialGrid[T]) FindNear(bounds mosaic.Rectangle) []T {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	set := map[T]struct{}{}
	minPoint, maxPoint := bounds.MinPoint(), bounds.MaxPoint()
	xMinIndex, yMinIndex := sg.Location(minPoint.X, minPoint.Y)
	xMaxIndex, yMaxIndex := sg.Location(maxPoint.X, maxPoint.Y)

	for x, xn := xMinIndex, xMaxIndex; x <= xn; x++ {
		for y, yn := yMinIndex, yMaxIndex; y <= yn; y++ {
			for _, item := range sg.Nodes[x][y].Items {
				set[item.value] = struct{}{}
			}
		}
	}

	return maps.Keys(set)
}

func (sg *SpatialGrid[T]) Drop() {
	sg.nodesMu.Lock()
	defer sg.nodesMu.Unlock()

	nodes := make([][]spatialGridNode[T], sg.SizeX)
	for iX := range nodes {
		nodes[iX] = make([]spatialGridNode[T], sg.SizeY)
		for iY := range sg.SizeY {
			nodes[iX][iY] = newSpatialGridNode[T](
				iX,
				iY,
				mosaic.NewRectangle(
					mosaic.NewVector(
						(float64(iX)*sg.ChunkSize)+sg.ChunkSize/2,
						(float64(iY)*sg.ChunkSize)+sg.ChunkSize/2,
					),
					sg.ChunkSize,
					sg.ChunkSize,
				),
			)
		}
	}

	sg.Nodes = nodes
	sg.itemCount = 0
}

func (sg *SpatialGrid[T]) Location(x, y float64) (xIndex, yIndex int) {
	xIndex = int((x / sg.ChunkSize))
	yIndex = int((y / sg.ChunkSize))

	if xIndex < 0 {
		xIndex = 0
	} else if xIndex > sg.SizeX-1 {
		xIndex = sg.SizeX - 1
	}

	if yIndex < 0 {
		yIndex = 0
	} else if yIndex > sg.SizeY-1 {
		yIndex = sg.SizeY - 1
	}

	return xIndex, yIndex
}

func (sg *SpatialGrid[T]) GetItemsAtLocation(x, y int) []T {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	return sg.Nodes[x][y].Values()
}

func (sg *SpatialGrid[T]) GetLocationWeight(x, y int) float64 {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	return sg.Nodes[x][y].weight
}

func (sg *SpatialGrid[T]) NodeAtPosition(x, y float64) spatialGridNode[T] {
	xIndex, yIndex := sg.Location(x, y)
	return sg.Node(xIndex, yIndex)
}

func (sg *SpatialGrid[T]) Node(x, y int) spatialGridNode[T] {
	return spatialGridNode[T]{
		Items:  sg.Nodes[x][y].Items,
		x:      sg.Nodes[x][y].x,
		y:      sg.Nodes[x][y].y,
		bounds: sg.Nodes[x][y].bounds,
		weight: sg.Nodes[x][y].weight,
	}
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

		edges = append(edges, sg.Node(nextX, nextY))
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

	start := sg.NodeAtPosition(x, y)

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

			err = process(currentNode.Values())
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

func (sg *SpatialGrid[T]) WeightedSearch(start, end mosaic.Vector, maxDepth int) ([]mosaic.Vector, error) {
	sg.nodesMu.RLock()
	defer sg.nodesMu.RUnlock()

	heuristic := func(from, to spatialGridNode[T]) float64 {
		return math.Abs(float64(from.x-to.x)) + math.Abs(float64(from.y-to.y))
	}

	type index struct {
		X int
		Y int
	}

	startNode := sg.NodeAtPosition(start.X, start.Y)
	endNode := sg.NodeAtPosition(end.X, end.Y)

	cameFrom := map[index]spatialGridNode[T]{}
	cameFrom[index{startNode.x, startNode.y}] = startNode
	costs := map[index]float64{}
	costs[index{startNode.x, startNode.y}] = 0

	pq := caravan.NewPQ[spatialGridNode[T]](true)
	pq.Enqueue(startNode, 0)

	currentDepth := 0
	for pq.Len() > 0 {
		if currentDepth > maxDepth {
			return []mosaic.Vector{}, ErrMaxDepthReached
		}

		currentNode, err := pq.Dequeue()
		if err != nil {
			return []mosaic.Vector{}, err
		}

		if currentNode.x == endNode.x && currentNode.y == endNode.y {
			break
		}

		edges := sg.Edges(currentNode)
		for i := 0; i < len(edges); i++ {
			newCost := costs[index{currentNode.x, currentNode.y}] + edges[i].weight
			if math.IsInf(newCost, 1) {
				continue
			}

			edgeCost, ok := costs[index{edges[i].x, edges[i].y}]
			if ok && newCost >= edgeCost {
				continue
			}

			costs[index{edges[i].x, edges[i].y}] = newCost
			node := spatialGridNode[T]{
				Items:  edges[i].Items,
				x:      edges[i].x,
				y:      edges[i].y,
				bounds: edges[i].bounds,
				weight: edges[i].weight,
			}
			priority := newCost + heuristic(edges[i], endNode)
			pq.Enqueue(node, priority)
			cameFrom[index{edges[i].x, edges[i].y}] = currentNode
		}
		currentDepth++
	}

	pathNodes := []spatialGridNode[T]{}
	currentNode := endNode
	_, ok := cameFrom[index{endNode.x, endNode.y}]
	if !ok {
		return []mosaic.Vector{}, ErrPathNotFound
	}

	for (index{currentNode.x, currentNode.y} != index{startNode.x, startNode.y}) {
		pathNodes = append(pathNodes, currentNode)
		currentNode = cameFrom[index{currentNode.x, currentNode.y}]
	}

	pathNodes = append(pathNodes, startNode)
	path := make([]mosaic.Vector, len(pathNodes))
	for i := len(pathNodes) - 1; i >= 0; i-- {
		path[len(pathNodes)-1-i] = mosaic.NewVector(
			(float64(pathNodes[i].x)*sg.ChunkSize)+sg.ChunkSize/2,
			(float64(pathNodes[i].y)*sg.ChunkSize)+sg.ChunkSize/2,
		)
	}

	return path, nil
}

func newSpatialGridNode[T comparable](x, y int, bounds mosaic.Rectangle) spatialGridNode[T] {
	return spatialGridNode[T]{
		Items:  make([]spatialGridNodeItem[T], 0, 512),
		x:      x,
		y:      y,
		bounds: bounds,
		weight: 0,
	}
}

func (sgn spatialGridNode[T]) Values() []T {
	values := make([]T, len(sgn.Items))
	for i := 0; i < len(values); i++ {
		values[i] = sgn.Items[i].value
	}

	return values
}

func (sgn spatialGridNode[T]) Insert(item T, bounds mosaic.Rectangle, multiplier float64) spatialGridNode[T] {
	weight := sgn.bounds.AreaOfOverlap(bounds) * multiplier

	sgn.Items = append(
		sgn.Items,
		newSpatialGridNodeItem(item, bounds, weight, multiplier),
	)
	sgn.weight += weight

	return sgn
}

func (sgn spatialGridNode[T]) Delete(item T) spatialGridNode[T] {
	for i := 0; i < len(sgn.Items); i++ {
		if sgn.Items[i].value != item {
			continue
		}
		sgn.weight = sgn.weight - sgn.Items[i].weight
		sgn.Items[i] = sgn.Items[len(sgn.Items)-1]
		sgn.Items = sgn.Items[:len(sgn.Items)-1]
	}

	return sgn
}

func newSpatialGridNodeItem[T comparable](value T, bounds mosaic.Rectangle, weight float64, multiplier float64) spatialGridNodeItem[T] {
	return spatialGridNodeItem[T]{
		value:      value,
		bounds:     bounds,
		multiplier: multiplier,
		weight:     weight,
	}
}
