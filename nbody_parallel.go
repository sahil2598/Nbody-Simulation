package main

import "math"
import "math/rand"
import "os"
import "strconv"
import "fmt"
import "sync"
import "time"

type Particle struct {
    x, y float64
    vx, vy float64
    node *TreeNode
}

type TreeNode struct
{
    particle *Particle
	totalMass float64
    lb, rb, db, ub float64
    child [4]*TreeNode
	mutex sync.Mutex
}

type Barrier struct {
    mutex *sync.Mutex
    cond *sync.Cond
    counter int
	threadCount int
}

var SOFTENING float64
var dt float64
var theta float64
func createNode(parent *TreeNode, childNumber int) *TreeNode {
    var newNode TreeNode
    newNode.particle = nil
    newNode.totalMass = 0

    switch childNumber {
    case 0: /* upper left square */
        newNode.lb = parent.lb
        newNode.rb = (parent.lb + parent.rb) / 2.0
        newNode.ub = parent.ub
        newNode.db = (parent.ub + parent.db) / 2.0
    case 1: /* upper right square */
        newNode.lb = (parent.lb + parent.rb) / 2.0
        newNode.rb = parent.rb
        newNode.ub = parent.ub
        newNode.db = (parent.ub + parent.db) / 2.0
    case 2: /* lower left square */
        newNode.lb = parent.lb
        newNode.rb = (parent.lb + parent.rb) / 2.0
        newNode.ub = (parent.ub + parent.db) / 2.0
        newNode.db = parent.db
    case 3: /* lower right square */
        newNode.lb = (parent.lb + parent.rb) / 2.0
        newNode.rb = parent.rb
        newNode.ub = (parent.ub + parent.db) / 2.0
        newNode.db = parent.db
    }

    return &newNode
}

func whichChildContains(t *TreeNode, p *Particle) *TreeNode {
    for i := 0; i < 4; i++ {
        temp := t.child[i]
        if p.x >= temp.lb && p.x <= temp.rb && p.y >= temp.db && p.y <= temp.ub {
            return t.child[i]
        }
    }

    return nil
}

func isLeaf(t *TreeNode) bool {
    for i := 0; i < 4; i++ {
        if t.child[i] != nil {
            return false
		}
    }

    return true
}

func treeInsert(t *TreeNode, p *Particle) {
    var temp *TreeNode
	t.mutex.Lock()
    if !isLeaf(t) {	/* internal node */
        temp = whichChildContains(t, p)
        treeInsert(temp, p)
    } else if t.particle != nil {		/* non-empty leaf node */
        for i := 0; i < 4; i++ {
            t.child[i] = createNode(t, i)
        }

        parentParticle := t.particle
        t.particle = nil
        temp = whichChildContains(t, parentParticle) 	/* assign parent particle to one of the child nodes */
        treeInsert(temp, parentParticle)

        temp = whichChildContains(t, p) /* insert particle p */
        treeInsert(temp, p)
    } else {		/* empty leaf node */
        t.particle = p
        p.node = t
    }
	t.mutex.Unlock()
    t.totalMass++ /* increment mass for parent and leaf nodes */
}

func calcCenterOfMass(node **TreeNode) {
    t := *node
    if t.particle != nil || t.totalMass <= 1 { /* do not calculate for leaf nodes */
        return
	}

    x1 := 0.0
	y1 := 0.0
    for i := 0; i < 4; i++ {
        temp := t.child[i]
        if temp != nil && temp.totalMass != 0 {
            p := temp.particle
            mass := temp.totalMass
            x1 += mass * p.x
            y1 += mass * p.y
        }
    }

	mass := t.totalMass
    var p Particle
    p.x = x1 / mass
    p.y = y1 / mass
    (*t).particle = &p
}

func populateCenterOfMass(t *TreeNode) { /* calculate center of mass for all internal nodes in bottom up fashion */
    if t == nil {
        return
	}
    for i := 0; i < 4; i++ {
        populateCenterOfMass(t.child[i])
	}

    calcCenterOfMass(&t)
}

func calcForce(particle1 **Particle, p2 *Particle, totalMass float64) {
    p1 := *particle1
    dx := p2.x - p1.x
    dy := p2.y - p1.y
    distSqr := dx * dx + dy * dy + SOFTENING
    invDist := 1.0 / math.Sqrt(distSqr)
    invDist3 := invDist * invDist * invDist

    Fx := dx * invDist3
    Fy := dy * invDist3

    massConstant := totalMass * dt
    p1.vx += massConstant * Fx
    p1.vy += massConstant * Fy
}

func isValid(t1 *TreeNode, p2 *Particle) bool {
    p1 := t1.particle
    dx := p1.x - p2.x
    dy := p1.y - p2.y
    distSqr := dx * dx + dy * dy + SOFTENING
    d := math.Sqrt(distSqr)
    s := math.Abs(t1.lb - t1.rb)
    ratio := s / d
    return ratio < theta
}

func computeNodeForce(curr *TreeNode, t *TreeNode) {
    if curr == nil || curr.totalMass == 0 {
        return
	}

    if curr.totalMass == 1 {		/* in case of leaf node, calculate force between the 2 particles */
        calcForce(&t.particle, curr.particle, 1)
    } else {
        if isValid(curr, t.particle) {	/* check if center of mass can be used for force calculation */
            calcForce(&t.particle, curr.particle, curr.totalMass)
		} else {	/* otherwise go down to children nodes */
            for i := 0; i < 4; i++ {
                computeNodeForce(curr.child[i], t)
			}
        }
    }
}

func traverseTree(t *TreeNode, root *TreeNode) {
    if t == nil {
        return
	}

    if t.totalMass == 1 {
        computeNodeForce(root, t) /* calculate force on particle if leaf node */
    } else {
        for i := 0; i < 4; i++ {
            temp := t.child[i]
            if temp != nil && temp.totalMass != 0 {
                traverseTree(temp, root)
			}
        }
    }
}

func max(a float64, b float64) float64 {
    if a > b {
        return a
    }
    return b
}

func min(a float64, b float64) float64 {
    if a < b {
        return a
    }
    return b
}

func randInit(data []Particle, n int) {
    r := rand.New(rand.NewSource(99))
    for i := 0; i < n; i++ {
		data[i].x = r.Float64()
		data[i].y = r.Float64()
		data[i].vx = r.Float64()
		data[i].vy = r.Float64()
    }
}

func (b *Barrier) barrierSync() {
    b.mutex.Lock()
    b.counter++
    if b.counter < b.threadCount {
        b.cond.Wait()
    } else {
        b.cond.Broadcast()
    }
    b.mutex.Unlock()
}

func performAllTasks(root *TreeNode, p []Particle, start int, end int, threadNum int, b1 *Barrier, b2 *Barrier, b3 *Barrier, wg *sync.WaitGroup) {
    for i := start; i < end; i++ {
        treeInsert(root, &p[i])
    }
    b1.barrierSync()
    if threadNum == 0 {
        populateCenterOfMass(root)
    }
    b2.barrierSync()
    for i := start; i < end; i++ {
        computeNodeForce(root, p[i].node)
    }
    b3.barrierSync()
    for i := start; i < end; i++ {
        p[i].x += p[i].vx * dt
        p[i].y += p[i].vy * dt
    }
    wg.Done()
}


func main() {
	nParticles := 3000
    nIters := 200
    SOFTENING = 1e-9
    dt = 0.01
    theta = 0.5

    if len(os.Args) > 1 {
        nParticles, _ = strconv.Atoi(os.Args[1])
	}

    if len(os.Args) > 2 {
        nIters, _ = strconv.Atoi(os.Args[2])
	}

    p := make([]Particle, nParticles)
    randInit(p, nParticles)

    /* to perform testing with circle, comment above 3 lines and uncomment these lines
    Particle *p = malloc(nParticles * sizeof(Particle));
    getCircle(p, nParticles);
    */

    datafile, _ := os.Create("particles_parallel.dat")
	content := fmt.Sprintf("%d %d %d\n", nParticles, nIters, 0)
	_, _ = datafile.WriteString(content)

    max_limit := 0.0
	min_limit := 0.0

    startTime := time.Now()
    for iter := 1; iter <= nIters; iter++ {
        fmt.Printf("Iteration: %d\n", iter);
        for i := 0; i < nParticles; i++ {
			content = fmt.Sprintf("%f %f \n", p[i].x, p[i].y)
            _, _ = datafile.WriteString(content)
            max_limit = max(max_limit, max(p[i].x, p[i].y))
            min_limit = min(min_limit, min(p[i].x, p[i].y))
        }

        max_limit++
        min_limit--

        var root TreeNode
        for i := 0; i < 4; i++ {
            root.child[i] = nil
        }
        root.particle = nil
        root.totalMass = 0
        root.lb = min_limit
        root.db = min_limit
        root.rb = max_limit
        root.ub = max_limit

		nThreads := 4
		particlesPerThread := int(math.Ceil(float64(nParticles) / float64(nThreads)))

        var mutex1 sync.Mutex
        var mutex2 sync.Mutex
        var mutex3 sync.Mutex
        cond1 := sync.NewCond(&mutex1)
        cond2 := sync.NewCond(&mutex2)
        cond3 := sync.NewCond(&mutex3)
        b1 := Barrier{mutex: &mutex1, cond: cond1, counter: 0, threadCount: nThreads}
        b2 := Barrier{mutex: &mutex2, cond: cond2, counter: 0, threadCount: nThreads}
        b3 := Barrier{mutex: &mutex3, cond: cond3, counter: 0, threadCount: nThreads}

        var wg sync.WaitGroup
        for i:= 0; i < nThreads; i++ {
            start := i * particlesPerThread
            end := start + particlesPerThread
            if end > nParticles {
                end = nParticles
            }
            wg.Add(1)
            go performAllTasks(&root, p, start, end, i, &b1, &b2, &b3, &wg)
        }
        wg.Wait()
    }
    endTime := time.Since(startTime).Seconds()
    fmt.Printf("Total time: %.15f\n", endTime)
}