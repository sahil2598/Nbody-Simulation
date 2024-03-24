package main

import (
	"proj3/execution"
	"os"
	"strconv"
	"fmt"
	"time"
	"proj3/nbody"
)

func main() {
	nParticles := 3000
    nIterations := 200

	if len(os.Args) > 1 {
        nParticles, _ = strconv.Atoi(os.Args[1])
	}
    if len(os.Args) > 2 {
        nIterations, _ = strconv.Atoi(os.Args[2])
	}

	execType := "s"
	if len(os.Args) > 3 {
        execType = os.Args[3]
	}

	nThreads := 1
	if execType != "s" {
		nThreads, _ = strconv.Atoi(os.Args[4])
	}

	particleArray := nbody.CreateParticleArray(nParticles)

	//Comment above line and uncomment below line for circular arrangement of particles
	// particleArray := nbody.GetCircle(nParticles)

    datafile, _ := os.Create("output/particles_" + execType + ".dat")
	content := fmt.Sprintf("%d %d %d\n", nParticles, nIterations, 0)
	_, _ = datafile.WriteString(content)

    startTime := time.Now()
    for iter := 1; iter <= nIterations; iter++ {
        fmt.Printf("Iteration: %d\n", iter)
		min_limit, max_limit := nbody.WriteToFile(particleArray, execType)
        root := nbody.InitRoot(min_limit, max_limit)

		switch execType {
		case "s":
			execution.RunSequential(root, particleArray)
		case "p":
			execution.RunParallel(root, particleArray, nThreads)
		case "w":
			execution.RunWorkSteal(root, particleArray, nThreads)
		}
    }
    endTime := time.Since(startTime).Seconds()
    fmt.Printf("Total time: %.15f\n", endTime)
}