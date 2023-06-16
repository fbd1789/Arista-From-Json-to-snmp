# intf_tc_queue_counters

## OID Structure

```
+--IntfTcQueueTable(1)
|  |
|  +--IntfTcQueueEntry(1)
|  |  |  Index: ifIndex, trafficClass
|  |  |
|  |  +-- String IntfTcQueueIndex(1)
|  |  |
|  |  +-- String IntfTcQueueName(2)
|  |  |
|  |  +-- Counter64 IntfTcQueueEnqueuedBytes(3)
|  |  |
|  |  +-- Counter64 IntfTcQueueEnqueuedPackets(4)
|  |  |
|  |  +-- Counter64 IntfTcQueueDroppedBytes(5)
|  |  |
|  |  +-- Counter64 IntfTcQueueDroppedPackets(6)

```

## Build

```
GOOS=linux go build -o intf_tc_queue_counters .
```

Note: the base OID can be changed by editing the source file

## Install

```
scp intf_tc_queue_counters admin@switch:/mnt/flash/
```

## Usage

### Debug/Testing

```bash
$ /mnt/flash/intf_tc_queue_counters
PING
```

_Output_

```
PONG
```

### GetNext

```
getnext
1.3.6.1.4.1.30065.4.224
```

_Output_

```
1.3.6.1.4.1.30065.4.224.1.1001.0
STRING
1001.0
```

#### Quit

```
^C
```

### Configure SNMP extension

```
configure
snmp-server extension .1.3.6.1.4.1.30065.4.224 flash:/intf_tc_queue_counters
```
