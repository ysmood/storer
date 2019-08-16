# bucket

This bucket lib is dependency free, you can use whatever backend you want.
This lib uses the [byframe](https://github.com/ysmood/byframe)
to prevent prefix conflict when dealing with large number of buckets.

Key conflict example, numbers are hex based.

```txt
|     whole key       |
| =================== |
| prefix   | key      |
| -------- | -------- |
| 00       | 01       |
| 01       | 01       | mark_a
| .        | .        |
| .        | .        |
| .        | .        |
| 01 01    | 01       | mark_b
| 01 02    | 01       |
```

As you can see the mark_a's whole-key is the prefix of mark_b.
So when you do prefix search for "0x01 01" mark_a will also be in the result.

The key is randomly passed from user, so it can be any content. Therefore the
design to prevent the conflict should be based on the prefix encoding itself.
If we can make sure the short prefix won't be the prefix of longer prefix then
the problem will be solved. [byframe](https://github.com/ysmood/byframe) is used to
enables it.

Here's an example for 200 buckets.

We insert data by these pseudo code:

```go
one = newBucket("bucket-01")

one.Set("key01", "value01")
one.Set("key02", "value02")

two = newBucket("bucket-02")
two.Set("key03", "value03")

...

n = newBucket("bucket-200")
n.Set("key200", "value200")
```

The pseudo layout of the db, numbers are hex based:

```txt
| key             |   value
| =============== | ============
| 00              | 2
| 00  bucket-01   | 01
| 00  bucket-02   | 02
| .
| .
| .
| 00  bucket-200  | C81 // not the hex of 200, encoded by byframe lib
| 01  key01       | value01
| 01  key02       | value02
| 02  key03       | value03
| C81 key200      | value200
```
