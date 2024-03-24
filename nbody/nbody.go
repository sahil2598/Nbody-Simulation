package nbody

import "math"
import "math/rand"
import "fmt"
import "os"
import "sync"

const (
    SOFTENING = 1e-9
    dt = 0.01
    theta = 0.5
)

type Particle struct {
    x, y float64
    vx, vy float64
    Node *TreeNode
}

type TreeNode struct
{
    particle *Particle
	totalMass float64
    lb, rb, db, ub float64
    child [4]*TreeNode
    mutex sync.Mutex
} 

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

/* check which child of parent node should hold the particle */
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

func TreeInsert(t *TreeNode, p *Particle, parallelFlag bool) {
    var temp *TreeNode
    if parallelFlag {
        t.mutex.Lock()
    }
    if !isLeaf(t) {	/* internal node */
        temp = whichChildContains(t, p)
        t.totalMass++
        if parallelFlag {
            t.mutex.Unlock()
        }
        TreeInsert(temp, p, parallelFlag)
    } else if t.particle != nil {		/* non-empty leaf node */
        for i := 0; i < 4; i++ {
            t.child[i] = createNode(t, i)
        }

        parentParticle := t.particle
        t.particle = nil
        temp = whichChildContains(t, parentParticle) 	/* assign parent particle to one of the child nodes */
        TreeInsert(temp, parentParticle, parallelFlag)

        t.totalMass++
        temp = whichChildContains(t, p) /* insert particle p */
        if parallelFlag {
            t.mutex.Unlock()    /* check if correct */
        }
        TreeInsert(temp, p, parallelFlag)
    } else {		/* empty leaf node */
        t.particle = p
        p.Node = t
        t.totalMass++
        if parallelFlag {
            t.mutex.Unlock()
        }
    }
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

/* calculate center of mass for all internal nodes in bottom up fashion */
func PopulateCenterOfMass(t *TreeNode) {
    if t == nil {
        return
	}
    for i := 0; i < 4; i++ {
        PopulateCenterOfMass(t.child[i])
	}

    calcCenterOfMass(&t)
}

/* calculate force applied on particle1 by particle2 */
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

/* check if center of mass can be used for force calculation */
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

/* compute total force applied on particle */
func ComputeNodeForce(curr *TreeNode, t *TreeNode) {
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
                ComputeNodeForce(curr.child[i], t)
			}
        }
    }
}

/* traverse tree to find all leaf nodes, and calculate force on each particle */
func TraverseTree(t *TreeNode, root *TreeNode) {
    if t == nil {
        return
	}

    if t.totalMass == 1 {
        ComputeNodeForce(root, t) /* calculate force on particle if leaf node */
    } else {
        for i := 0; i < 4; i++ {
            temp := t.child[i]
            if temp != nil && temp.totalMass != 0 {
                TraverseTree(temp, root)
			}
        }
    }
}

/* update position of particle after calculation of force */
func UpdatePosition(p *Particle) {
    p.x += p.vx * dt
    p.y += p.vy * dt
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

/* random initialization of particles in (0, 1) */
func randInit(data []Particle, n int) {
    r := rand.New(rand.NewSource(99))
    for i := 0; i < n; i++ {
		data[i].x = r.Float64()
		data[i].y = r.Float64()
		data[i].vx = r.Float64()
		data[i].vy = r.Float64()
    }
}

func CreateParticleArray(nParticles int) []Particle {
    particleArray := make([]Particle, nParticles)
    randInit(particleArray, nParticles)
    return particleArray
}

/* write particle positions to file */
func WriteToFile(particleArray []Particle, execType string) (float64, float64) {
    max_limit := 0.0
	min_limit := 0.0
    datafile, _ := os.OpenFile("output/particles_" + execType + ".dat", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    for i := 0; i < len(particleArray); i++ {
        content := fmt.Sprintf("%f %f \n", particleArray[i].x, particleArray[i].y)
        _, _ = datafile.WriteString(content)
        max_limit = max(max_limit, max(particleArray[i].x, particleArray[i].y))
        min_limit = min(min_limit, min(particleArray[i].x, particleArray[i].y))
    }
    max_limit++
    min_limit--

    return min_limit, max_limit
}

/* initialize root of quad tree */
func InitRoot(min_limit float64, max_limit float64) *TreeNode {
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

    return &root
}

/* get start and end index of particle array for goroutine */
func GetStartAndEnd(idx int, nParticles int, particlesPerThread int) (int, int) {
    start := idx * particlesPerThread
    end := start + particlesPerThread
    if end > nParticles {
        end = nParticles
    }
    return start, end
}

/* extraneous method to arrange particles in a circle */
func GetCircle(nParticles int) []Particle {
	p := make([]Particle, nParticles)
    radius := 1.0
    for i := 0; i < nParticles; i++ {
        angle := 2.0 * math.Pi * float64(i) / float64(nParticles)
        p[i].x = radius * math.Cos(angle)
        p[i].y = radius * math.Sin(angle)
        p[i].vx = 0
        p[i].vy = 0
    }
	return p
}