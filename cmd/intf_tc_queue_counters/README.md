# intf_tc_queue_counters

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
