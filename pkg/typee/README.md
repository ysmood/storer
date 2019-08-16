# typee

Seamless migration can be achieved by hook the decoding stage of data reading.
Only when we read the data from db the type matters, when we write the data to
the db, we can always use the latest type.

We only need upgrade, an downgrade should be converted to an upgrade, so that we
are able to track the whole history of the schema.

## How the migration works

The lib will basically create a block chain hash of the type and its history types.

For example:

```txt
T0 {
    Age int
}

T1 {
    Age string
}

T {
    Age int
}
```

The relationshop is `T0 -> T1 -> T`, the hash of each type will be like:

- T0: `{Age int}` -> `0x01`
- T1: `{Age int}{Age string}` -> `0x02`
- T : `{Age int}{Age string}{Age int}`  -> `0x03`

When encoded to json, the result will be like:

- T0: `0x01{Age: 10}`.
- T1: `0x02{Age: "10"}`.
- T : `0x03{Age: 10}`.

When decoding an item, the lib will check the version first, and call the corresponding
migrate methods to migrate data to the lastest schema.

Such as when encounter data `0x01{Age: 1}`, the lib will first migrate it to `0x02{Age: "1"}`
and then migrate it to `0x03{Age: 1}`. The migrated data will be written back to db,
so same item won't be migrated again.

Only the items that are read will be migrated, unused items won't be migrated.
Systems that require liveness can benifit from it.
