# go-grep

A grep like tool for a binary file


## Usage

```
$ bin-grep [OPTIONS] PATTERN [FILE...]
```

### Pattern

Pattern should be hex string. e.g. `"0xaabb"` or `"aabb"`
You can use `.` as any byte.

`"0x.bb"` will be matched `"aabb"`.

### Option

|Option|Description|
|------|-----------|
|`-c`| print a count of matched case|
|`-a`| print only an matched address|
|`-s uint`|skip n bytes|
|`-h`| show help |
|`-V`| show version|

## License

[Apache License v2.0](https://www.apache.org/licenses/LICENSE-2.0)