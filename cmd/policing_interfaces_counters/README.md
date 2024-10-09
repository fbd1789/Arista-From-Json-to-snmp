# policing interfaces counters


```

+-- policingInterfaceCounters(<derived>)
|   |
|   +-- ingressInterfacesTable(1)
|   |   |
|   |   +-- ingressInterfacesEntry(1)
|	|	|   | Index: [ifIndex]
|   |	|   |
|   |   |   +-- String ifName(1)
|   |   |   |
|   |   |   +-- String profile(2)
|   |   |   |
|   |   |   +-- String mode(3)
|   |   |   |
|   |   |   +-- Counter64 conformedPackets(4)
|   |   |   |
|   |   |   +-- Counter64 conformedBytes(5)
|   |   |   |
|   |   |   +-- Counter64 yellowPackets(6)
|   |   |   |
|   |   |   +-- Counter64 yellowBytes(7)
|   |   |   |
|   |   |   +-- Counter64 exceededPackets(8)
|   |   |   |
|   |   |   +-- Counter64 exceededBytes(9)
|   |   |
|   +-- egressInterfacesTable(2)
|   |   |
|   |   +-- egressInterfacesEntry(1)
|	|	|   | Index: [ifIndex]
|   |	|   |
|   |   |   +-- String ifName(1)
|   |   |   |
|   |   |   +-- String profile(2)
|   |   |   |
|   |   |   +-- String mode(3)
|   |   |   |
|   |   |   +-- Counter64 conformedPackets(4)
|   |   |   |
|   |   |   +-- Counter64 conformedBytes(5)
|   |   |   |
|   |   |   +-- Counter64 yellowPackets(6)
|   |   |   |
|   |   |   +-- Counter64 yellowBytes(7)
|   |   |   |
|   |   |   +-- Counter64 exceededPackets(8)
|   |   |   |
|   |   |   +-- Counter64 exceededBytes(9)

```