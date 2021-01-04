# Kagi-db
![Go](https://github.com/amohamed11/kagi/workflows/Go/badge.svg)  
Kagi is a lightweight key-value database built in Go.  
You can find my notes on b-tree & key-value store design at `notes/`.  

## Motivations
- Research more about database design & implementation
- Apply learned knowledge & theory from classes (db design, access methods, B-trees, etc) 
- Learn & practice Go (simultaneously reading [The Go Programming Language](https://www.gopl.io/)) 

## Functionalities
- [x] Save data to db  
- [x] Retreive data from db  
- [ ] Delete data from db  
- [ ] Create groupings as an umbrella for related data  

## Benchmarks
Benchmarks will be updated once Kagi is complete, and performance can be focused on.  
Although performance is not the main priority, I still aim to make Kagi reasonably performant.  
Ran using Go's testing library/tool. For benchmarking code check `bench_test.go`.  
| Operation       | time (ms) |
|-----------------|----------:|
| Set (1000 Keys) |      26.7 |
| Get (1000 Keys) |       2.1 |
