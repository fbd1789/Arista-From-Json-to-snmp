# bgp-vrf-all

# Structure

```
BaseOID: 1.3.6.1.4.1.8072.1.3.1.5

+--- cpuCountersQueue (11)
    |
    +--- egressQueuesSummaryTable(1)
    |   |
    |   +--- egressQueuesSummaryEntry(1)
    |       | Index: portId, destTypeId, queueId
    |       |
    |       +-- String port(1)
    |       |
    |       +-- String destType(2)
    |       |
    |       +-- Integer queue(3)
    |       |
    |       +-- Counter64 enqueuedPackets(4)
    |       |
    |       +-- Counter64 enqueuedBytes(5)
    |       |
    |       +-- Counter64 droppedPackets(6)
    |       |
    |       +-- Counter64 droppedBytes(7)
    |
    +--- ingressVoqsSummaryTable(2)
    |   |
    |   +--- ingressVoqsSummaryEntry(1)
    |       | Index: CoppClassId
    |       |
    |       +-- String coppClass(1)
    |       |
    |       +-- Counter64 enqueuedPackets(2)
    |       |
    |       +-- Counter64 enqueuedBytes(3)
    |       |
    |       +-- Counter64 droppedPackets(4)
    |       |
    |       +-- Counter64 droppedBytes(5)
```

# Build

```
GOOS=linux GOARCH=amd64 scpCGO_ENABLED=0 go build -o cpu_counters_queue .
```

# Installation

```
scp cpu_counters_queue admin@switch:/mnt/flash/
```

# Configuration

```
switch#configure
switch(config)#snmp-server extension 1.3.6.1.4.1.8072.1.3.1.5 flash:/cpu_counters_queue
```

# Verify

_Wait few seconds for the data to populate if there are large number of prefixes_

```
switch#show snmp mib walk 1.3.6.1.4.1.8072.1.3.1.5
```

