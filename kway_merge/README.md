How to run standalone merger:
============================

$ make

Sample output of test run (from my mac):
=======================================

```
gcc -I/usr/local/Cellar/icu4c/55.1/include  -L/usr/local/Cellar/icu4c/55.1/lib  min_heap.c collate_json.c kway_merge.c kway_merge_test.c -o output -licui18n -licuuc -licudata
./output
Merging 4 arrays each of size 1000000
mergeKArrays took 5.662287s
```
