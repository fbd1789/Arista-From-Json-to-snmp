# bgp_summary_vrf_all

```

+-- bgpSummaryVrfAll(<derived>)
|   |
|   +-- ipBgpSummaryVrfAllTable(1)
|   |   |
|   |   +-- ipBgpSummaryVrfAllEntry(1)
|	|	|   | Index: [neighbor][vrfName]
|   |	|   |
|   |   |   +-- String vrfName(1)
|   |   |   |
|   |   |   +-- IpAddress neighbor(2)
|   |   |   |
|   |   |   +-- Integer asn(3)
|   |   |   |
|   |   |   +-- TimeTicks uptime(4)
|   |   |   |
|   |   |   +-- Integer prefixReceived(5)
|   |   |   |
|   |   |   +-- Integer prefixAccepted(6)
|   |   |
|   +-- ipv6VgpSummaryVrfAllTable(2)
|   |   |
|   |   +-- ipv6VgpSummaryVrfAllEntry(1)
|	|	|   | Index: [neighbor][vrfName]
|   |	|   |
|   |   |   +-- String vrfName(1)
|   |   |   |
|   |   |   +-- IpAddress neighbor(2)
|   |   |   |
|   |   |   +-- Integer asn(3)
|   |   |   |
|   |   |   +-- TimeTicks uptime(4)
|   |   |   |
|   |   |   +-- Integer prefixReceived(5)
|   |   |   |
|   |   |   +-- Integer prefixAccepted(6)

```