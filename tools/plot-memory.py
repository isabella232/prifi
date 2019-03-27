#!/usr/bin/python3

import sys
from datetime import datetime, timedelta
import matplotlib.pyplot as plt
import numpy

xs = []

cpu = []
mem = []
vsz = []
rss = []

alloc = []
heapobjects = []
heapinuse = []
heapidle = []
heapsys = []
stacksys = []
sys2 = []

trustee_mem = []
client_mem = []

i = 0
with open('log', 'r') as f:
    for line in f:
        i += 1
        print(i)
        line = line.strip()
        if line.startswith('root') or "Relay Memory" in line or "HeapObjects" in line or "[BufferableRoundManager]" in line:
            if "Relay Memory" in line:
                roundID = line[line.find(' - Round')+9:]
                roundID = roundID[0:roundID.find(' ')]
                xs.append(int(roundID))
            if 'root' in line:
                parts = [x for x in line.split(' ') if x != ""]
                cpu.append(float(parts[2]))
                mem.append(float(parts[3]))
                vsz.append(float(parts[4]) / 1024)
                rss.append(float(parts[5]) / 1024)
            if "HeapObjects" in line:
                parts = [x for x in line.split(' ') if x != ""]
                print(parts)
                alloc.append(int(parts[2]))
                heapobjects.append(int(parts[6]))
                heapinuse.append(int(parts[10]))
                heapidle.append(int(parts[13]))
                heapsys.append(int(parts[16]))
                stacksys.append(int(parts[19]))
                sys2.append(int(parts[23]))

            if "[BufferableRoundManager]" in line:
                line = line[line.find('[BufferableRoundManager]')+25:]
                parts = [x for x in line.split(';') if x != ""]

                trustee_mem.append(float(parts[0].split(' ')[3])/1024/1024)
                client_mem.append(float(parts[1].strip().split(' ')[3])/1024/1024)


fig, ax = plt.subplots()
l1 = ax.plot(xs, cpu, label='% CPU')
l2 = ax.plot(xs, mem, label='% Memory')
ax.set_xlabel('Rounds')
ax.set_ylabel('% Busy')
plt.legend(loc='best')
plt.grid(linestyle='-', linewidth=1)

fig, ax = plt.subplots()
ax.plot(xs, vsz, label='Unix command: VSZ (Virtual memory, including swapped)')
ax.plot(xs, rss, label='Unix command: RSS (Total in-RAM, including stack/heap)')
ax.plot(xs, alloc, label='Go runtime: Total heap-allocated memory')
ax.plot(xs, sys2, label='Go runtime: Total obtained from OS')
ax.plot(xs, heapinuse, label='Go runtime: HeapInUse')
ax.plot(xs, heapidle, label='Go runtime: HeapIdle')
ax.plot(xs, heapsys, label='Go runtime: HeapSys')
#ax.plot(xs, trustee_mem, label='Trustee Memory')
#ax.plot(xs, client_mem, label='Client Memory')
ax.set_xlabel('Rounds')
ax.set_ylabel('Mb')
plt.legend(loc='best')
plt.grid(linestyle='-', linewidth=1)

fig, ax = plt.subplots()
ax.plot(xs, heapobjects, label='Go runtime: HeapObjects')
ax.plot(xs, trustee_mem, label='Trustee Memory')
ax.plot(xs, client_mem, label='Client Memory')
ax.set_xlabel('Rounds')
ax.set_ylabel('B')
plt.legend(loc='best')
plt.grid(linestyle='-', linewidth=1)

plt.show()