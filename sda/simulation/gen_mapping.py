import os

outFile="hosts_mapping_remote_trustees.toml"

nTrustees=5
nTrusteeMachines=3
nClients=100
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

print("Done")