# bgp-vrf-all

# Structure

```
BaseOID: 1.3.6.1.4.1.30065.4.226

+-- bgpVrfAll(1)
|   |
|   +-- bgpVrfAllTable(1)
|   |   |
|   |   +-- bgpVrfAllEntry(1)
|	|	|   | Index: [address][maskLen][nextHop][vrfName]
|   |	|   |
|   |   |   +-- String vrfName(1)
|   |   |   |
|   |   |   +-- IpAddress prefix(2)
|   |   |   |
|   |   |   +-- Integer prefixLen int(3)
|   |   |   |
|   |   |   +-- IpAddress nextHop metric(4)
```

# Build

```
GOOS=linux go build -o bgp-vrf-all-linux-amd64 .
```

# Installation

```
scp bgp-vrfs-all-linux-amd64 admin@switch:/mnt/flash/
```

# Configuration

```
switch#configure
switch(config)#snmp-server extension .1.3.6.1.4.1.30065.4.226 flash:/bgp-vrf-all-linux-amd64
```

# Verify

_Wait few minutes for the data to populate if there are large number of prefixes_

```
switch#show snmp mib walk 1.3.6.1.4.1.30065.4.226
```

