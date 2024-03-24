package execution

import (
	"math"
	"math/rand"
	"proj3/queue"
	"proj3/nbody"
	"sync"
	"sync/atomic"
)

func nbodyWorkSteal(root *nbody.TreeNode, particleArray []nbody.Particle, start int, end int, threadNum int, nThreads int32, b1 *Barrier, b2 *Barrier, b3 *Barrier, insertQueues []*queue.DEQueue, computeQueues []*queue.DEQueue, wg *sync.WaitGroup, insertCount *int32, computeCount *int32) {
	for {
		particleIdx := insertQueues[threadNum].PopBottom()
		if particleIdx == -1 {
			break
		}
        nbody.TreeInsert(root, &particleArray[particleIdx], true)
    }
	atomic.AddInt32(insertCount, 1)

	for *insertCount < nThreads {
		idx := rand.Int31n(nThreads)
		particleIdx := insertQueues[idx].PopTop()
		if particleIdx == -1 {
			continue
		} 
		nbody.TreeInsert(root, &particleArray[particleIdx], true)
	}

    b1.barrierSync()

    if threadNum == 0 {
        nbody.PopulateCenterOfMass(root)
    }

    b2.barrierSync()

    for {
		particleIdx := computeQueues[threadNum].PopBottom() 
		if particleIdx == -1 {
			break
		}
        nbody.ComputeNodeForce(root, particleArray[particleIdx].Node)
    }
	atomic.AddInt32(computeCount, 1)

	for *computeCount < nThreads {
		idx := rand.Int31n(nThreads)
		particleIdx := computeQueues[idx].PopTop()
		if particleIdx == -1 {
			continue
		} 
		nbody.ComputeNodeForce(root, particleArray[particleIdx].Node)
	}

    b3.barrierSync()

    for i := start; i < end; i++ {
        nbody.UpdatePosition(&particleArray[i])
    }
    wg.Done()
}

func RunWorkSteal(root *nbody.TreeNode, particleArray []nbody.Particle, nThreads int) {
	nParticles := len(particleArray)
    particlesPerThread := int(math.Ceil(float64(nParticles) / float64(nThreads)))

	var mutex1, mutex2, mutex3 sync.Mutex
	cond1 := sync.NewCond(&mutex1)
	cond2 := sync.NewCond(&mutex2)
	cond3 := sync.NewCond(&mutex3)
	b1 := Barrier{mutex: &mutex1, cond: cond1, counter: 0, threadCount: nThreads}
	b2 := Barrier{mutex: &mutex2, cond: cond2, counter: 0, threadCount: nThreads}
	b3 := Barrier{mutex: &mutex3, cond: cond3, counter: 0, threadCount: nThreads}

	insertQueues := make([]*queue.DEQueue, nThreads)
	computeQueues := make([]*queue.DEQueue, nThreads)

	var wg sync.WaitGroup
	for i:= 0; i < nThreads; i++ {
		start, end := nbody.GetStartAndEnd(i, nParticles, particlesPerThread)
		insertQueues[i] = queue.NewDEQueue(start, end)
		computeQueues[i] = queue.NewDEQueue(start, end)
	}

	var insertCount, computeCount int32 = 0, 0
	for i := 0; i < nThreads; i++ {
		start, end := nbody.GetStartAndEnd(i, nParticles, particlesPerThread)
		wg.Add(1)
		go nbodyWorkSteal(root, particleArray, start, end, i, int32(nThreads), &b1, &b2, &b3, insertQueues, computeQueues, &wg, &insertCount, &computeCount)
	}
	wg.Wait()
}