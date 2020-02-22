#!/usr/local/bin/python3

import os
import sys

outFile="hosts_mapping_remote_trustees.toml"

nTrustees=5
nTrusteeMachines=3
nClients=1000
nClientsMachines=5

with open(outFile, 'w') as f:
    f.write('''[[hosts]]
  ID = 0
  IP = "10.0.1.254"

''')

    for i in range(nTrustees):
        f.write('''[[hosts]]
  ID = {0}
  IP = "10.1.0.{1}"

'''.format(1+i, (i%nTrusteeMachines)+1))

    for i in range(nClients):
        f.write('''[[hosts]]
  ID = {0}
  IP = "10.0.1.{1}"

'''.format(1+nTrustees+i, (i%nClientsMachines)+1))

if(len(sys.argv)) != 2:
    print("Usage: gen_mapping.py N_TRUSTEES")
    sys.exit(1)

nTrustees = int(sys.argv[1])


with open(outFile, 'w') as f:
    f.write('''[[hosts]]
  ID = 0
  IP = "10.0.1.254"

''')

    for i in range(nTrustees):
        f.write('''[[hosts]]
  ID = {0}
  IP = "10.1.0.{1}"

'''.format(1+i, (i%nTrusteeMachines)+1))

    for i in range(nClients):
        f.write('''[[hosts]]
  ID = {0}
  IP = "10.0.1.{1}"

'''.format(1+nTrustees+i, (i%nClientsMachines)+1))

print("Written", outFile, "for", nTrustees, "trustees")