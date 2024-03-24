package execution

import (
    "proj3/nbody"
    "sync"
    "math"
)

type Barrier struct {
    mutex *sync.Mutex
    cond *sync.Cond
    counter int
    threadCount int
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

func nbodyParallel(root *nbody.TreeNode, p []nbody.Particle, start int, end int, threadNum int, b1 *Barrier, b2 *Barrier, b3 *Barrier, wg *sync.WaitGroup) {
    for i := start; i < end; i++ {
        nbody.TreeInsert(root, &p[i], true)
    }

    b1.barrierSync()

    if threadNum == 0 {
        nbody.PopulateCenterOfMass(root)
    }

    b2.barrierSync()

    for i := start; i < end; i++ {
        nbody.ComputeNodeForce(root, p[i].Node)
    }

    b3.barrierSync()

    for i := start; i < end; i++ {
        nbody.UpdatePosition(&p[i])
    }
    
    wg.Done()
}

func RunParallel(root *nbody.TreeNode, particleArray []nbody.Particle, nThreads int) {
    nParticles := len(particleArray)
    particlesPerThread := int(math.Ceil(float64(nParticles) / float64(nThreads)))

	var mutex1, mutex2, mutex3 sync.Mutex
	cond1 := sync.NewCond(&mutex1)
	cond2 := sync.NewCond(&mutex2)
	cond3 := sync.NewCond(&mutex3)
	b1 := Barrier{mutex: &mutex1, cond: cond1, counter: 0, threadCount: nThreads}
	b2 := Barrier{mutex: &mutex2, cond: cond2, counter: 0, threadCount: nThreads}
	b3 := Barrier{mutex: &mutex3, cond: cond3, counter: 0, threadCount: nThreads}

	var wg sync.WaitGroup
	for i:= 0; i < nThreads; i++ {
		start, end := nbody.GetStartAndEnd(i, nParticles, particlesPerThread)
		wg.Add(1)
		go nbodyParallel(root, particleArray, start, end, i, &b1, &b2, &b3, &wg)
	}
	wg.Wait()
}
