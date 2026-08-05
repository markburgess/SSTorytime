#!/usr/bin/python3
#######################################################
# SST
#######################################################

import SSTorytime as SST

#######################################################
# Main
#######################################################

ok,sst = SST.Open("sstoryline","sst_1234","sstoryline","localhost")

if not ok:
    print("Couldn't open database")
    exit()

print("------- Define and retrieve notes with link  --------")

v1 = SST.Vertex(sst,"first node","examples chapter")
v2 = SST.Vertex(sst,"second node","examples chapter")

context = ['dunnum', 'cotton', 'pickin','lumberjack']

SST.Edge(sst,v1,"then",v2,context,1.0)

fetch1 = SST.GetDBNodeByNodePtr(sst,v1)
print("RESULT v1:",fetch1)

fetch2 = SST.GetDBNodeByNodePtr(sst,v2)
print("RESULT v2:",fetch2)

# Access class and instance variables

print("\n------- Now simple search for paths in examples --------")

leadsto = 1
contains = 2
express = 3
result_limit = 30

# Simplest cone search

link_paths,dim = SST.GetFwdPathsAsLinks(sst,"(4,1)",leadsto,result_limit,100)

if dim > 0:
    for path in link_paths:
        print("Path: ",end="")
        for lnk in path:
            node = SST.GetDBNodeByNodePtr(sst,lnk[3])
            print(lnk[3],"=",node[0],end=", ")
        print("\n")

print("\n------- Now more sopisticated for paths in examples --------")

# All singing, all dancing cone search

context = ['promises']
startset = [ '(4,6)' ]

super_paths,sdim = SST.GetEntireNCConePathsAsLinks(sst,"fwd",startset,10,"",context,result_limit)

if sdim > 0:
    for path in super_paths:
        print("Path: ",end="")
        for lnk in path:
            node = SST.GetDBNodeByNodePtr(sst,lnk[3])
            print(lnk[3],"=",node[0],end=", ")
        print("\n")

SST.Close(sst)
