
USAGE:
  Populate .pcache for users willing to optimize file lookup time
  with slow hard drives and random access to a deep nested objects 

It needs an argument of this list of a drives

```
./populate-hardlinks /mnt/minio-drives*/
```

By default, it only creates .pcache for files under which has more than three prefixes.

