#!/usr/bin/python3
#######################################################
# SST
#######################################################

import SSTorytime as SST

#######################################################
# Main
#######################################################

ok,ctx = SST.Open("sstoryline","sst_1234","sstoryline","localhost")

if not ok:
    print("Couldn't open database")
    exit()

print("------- Define and retrieve notes with link  --------")

v1 = SST.Vertex(ctx,"first node","examples chapter")
v2 = SST.Vertex(ctx,"second node","examples chapter")

context = ['dunnum', 'cotton', 'pickin','lumberjack']

SST.Edge(ctx,v1,"then",v2,context,1.0)

fetch1 = SST.GetDBNodeByNodePtr(ctx,v1)
print("RESULT v1:",fetch1)

fetch2 = SST.GetDBNodeByNodePtr(ctx,v2)
print("RESULT v2:",fetch2)

# Access class and instance variables

print("------- Now simple search for paths in examples --------")

leadsto = 1
result_limit = 30

# Simplest cone search

link_paths,dim = SST.GetFwdPathsAsLinks(ctx,"(1,0)",leadsto,result_limit,100)

for path in link_paths:
    print("Path: ",end="")
    for lnk in path:
        node = SST.GetDBNodeByNodePtr(ctx,lnk[3])
        print(lnk[3],"=",node[0],end=", ")
    print("\n")

print("------- Now more sopisticated for paths in examples --------")

# All singing, all dancing cone search

context = ['path']
startset = [ '(1,0)' ]

super_paths,sdim = SST.GetEntireNCSuperConePathsAsLinks(ctx,"fwd",startset,10,"multi",context,result_limit)

for path in super_paths:
    print("Path: ",end="")
    for lnk in path:
        node = SST.GetDBNodeByNodePtr(ctx,lnk[3])
        print(lnk[3],"=",node[0],end=", ")
    print("\n")


SST.Close(ctx)
