package execution

import "proj3/nbody"

func RunSequential(root *nbody.TreeNode, particleArray []nbody.Particle) {
	nParticles := len(particleArray)
	for i := 0; i < nParticles; i++ {
		nbody.TreeInsert(root, &particleArray[i], false)
	}

	nbody.PopulateCenterOfMass(root)
	nbody.TraverseTree(root, root)

	for i := 0; i < nParticles; i++ {
		nbody.UpdatePosition(&particleArray[i], )
	}
}