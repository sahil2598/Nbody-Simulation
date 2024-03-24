import matplotlib.pyplot as plt
from matplotlib.animation import FuncAnimation, PillowWriter
import sys
import numpy as np

if __name__ == "__main__":
    num_particles = int(sys.argv[1])
    num_iters = int(sys.argv[2])
    exec_type = sys.argv[3]

    fig, ax = plt.subplots(1, 1)
    lines = np.loadtxt('output/particles_' + exec_type + '.dat', skiprows=1, dtype=float)
    max_limit = int(lines.max()) + 1
    min_limit = int(lines.min()) + 1

    def animate(frame):
        start = frame * num_particles
        end = start + num_particles
        line_slice = lines[start:end]
        ax.clear()
        ax.title.set_text('Particle Positions')
        ax.set_xlim(min_limit, max_limit)
        ax.set_ylim(min_limit, max_limit)

        for line in line_slice:
            ax.scatter(line[0], line[1])

    ani = FuncAnimation(fig, animate, frames=num_iters, interval=10, repeat=False)
    writer = PillowWriter(fps=15)
    ani.save('output/nbody.gif', writer=writer)
    plt.close()