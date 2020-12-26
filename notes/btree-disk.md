# B+ Trees on Disk
## Structure
non-accurate sketch
<img src="./images/page-node.png" width="400">  

**Header:**  
- flags: describes node content  
  * isRoot: true if the node is root node  
  * isDeleted: true if node was deleted indicating that cleanup is needed  
- offSet: the position of node  
- parentOffset: position of parent node  
- numChildren: number of children nodes  
- childOffset: offset of the first child node
  
**Entry:**  
- currSize: current size of the value  
- freeOffset: offset or where the free space starts  
- value: string value of the leaf  
