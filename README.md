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

# N-body Simulation with Barnes Hut Algorithm

This project numerically approximates the evolution of a system of bodies in which each body continuously interacts with every other body through gravitational forces. It is particularly useful for simulating interactions between individual objects, such as stars or planets, in a two-dimensional space.

## Overview

An **N-body simulation** involves calculating gravitational forces between multiple bodies. The force between two bodies is given by:

\[ F = \frac{G \cdot m1 \cdot m2 \cdot R}{r^3} \]

where:
- **G** is the universal gravitational constant.
- **m1** and **m2** are the masses of the bodies.
- **R** is the position vector (x, y) between the bodies.
- **r** is the scalar distance between the bodies.

## Barnes Hut Algorithm

The **Barnes Hut algorithm** optimizes the simulation by reducing the time complexity to \(O(n \log n)\) through a hierarchical scheme of grouping bodies using a quad-tree data structure. The key steps are:

1. **Insertion of particles into the quad-tree**:
   - Recursively divide the space into quadrants until each subdivision contains 0 or 1 bodies.
2. **Calculation of the center-of-mass for all internal nodes**:
   - Compute the combined mass and center-of-mass of bodies within each quadrant.
3. **Force calculation for all particles**:
   - Calculate gravitational forces using the center-of-mass of nearby groups.
4. **Update of particle positions based on the calculated forces**:
   - Adjust positions of bodies according to the computed forces.

## Implementation

This project implements the Barnes Hut algorithm in three variants:
- **Serial (s)**: Single-threaded execution.
- **Parallel (p)**: Multi-threaded execution for improved performance.
- **Work stealing (w)**: Advanced parallel execution with dynamic load balancing.

## Execution

```bash
go run main.go <num_particles> <num_iterations>  s/p/w (type of execution)
```
