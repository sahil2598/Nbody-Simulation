An N-body simulation numerically approximates the evolution of a system of bodies in which each body continuously interacts with every other body through gravitational forces. These calculations are used in situations where interactions between individual objects, such as stars or planets, are important to the evolution of the system. This project focuses on the production of this simulation in two dimensions. The gravitational force between 2 bodies is calculated using the formula:
F = Gm1m2R/r^3, where G = universal constant, m1 and m2 are the mass of the bodies, R = position vector (x, y), r = scalar distance between the bodies

The Barnes Hut algorithm reduces the operations and time complexity to O(n logn) through a clever scheme for grouping together bodies that are sufficiently nearby. The method involves recursively dividing the set of bodies into groups using a quad-tree data structure. The topmost node represents the whole space, with its four children representing the four quadrants of the space. As shown in the diagram, the space is recursively subdivided into quadrants until each subdivision contains 0 or 1 bodies.

The algorithm contains 4 major steps:
1. Insertion of particles into the quad tree.
2. Calculation of the center-of-mass for all internal nodes.
3. Force calculation for all particles.
4. Update of the particle positions based on the calculated forces

This project implements the Barnes Hut algorithm serially, in parallel and with a work stealing refinement.

**Command for execution:** go run main.go <num_particles> <num_iterations>  s/p/w (type of execution)
