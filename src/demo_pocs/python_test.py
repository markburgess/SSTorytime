#!/usr/bin/python3
#######################################################
# SST
#######################################################

import psycopg2

#######################################################

class Node:

    # Class variable
    # species = "Canine"

    def __init__(self,name,chapter):
        self.S = name
        self.L = len(name)
        self.Chap = chapter
        self.NPtr = "undefined"
        self.Im3 = []
        self.Im2 = []
        self.Im1 = []
        self.In0 = []
        self.Il1 = []
        self.Ic2 = []
        self.Ie3 = []

    def IdempDBAddNode(name,chapter):
        node = Node(name,chapter)
        node.NPtr = "(1,2)"
        return node

#######################################################
    
class SST:

    def __init__(self):
        return

    def Open(dbuser,dbpwd,dbname,dbhost):

        try:
            conn = psycopg2.connect(database=dbname,user=dbuser,password=dbpwd,host=dbhost,port=5432)
        except:
            print("failed")
            return False,conn
        
        return True,conn

    #
    
    def Close(conn):
        conn.close()

    #

    def SQLEscape(s):
        return s.replace("'","\\'")

    #
    
    def NChannel(s):
        spc = s.count(" ")
        match spc:
            case 0:
                return 1
            case 1:
                return 2
            case 2:
                return 3

        l = len(s)
        if l < 128:
            return 4
        if l < 1024:
            return 5
        else:
            return 6
        
    #
        
    def GetDBNodeByNodePtr(conn,nptr):
        curs = conn.cursor()
        curs.execute("SELECT S,L,Chap,NPtr,Im3,Im2,Im1,In0,Il1,Ic2,Ie3 FROM Node where nptr='"+nptr+"'::NodePtr")
        pg_rows = curs.fetchall()
        conn.commit()
        return pg_rows[0]

    #

    def GetDBArrowsWithArrowName(conn,arrow):
        curs = conn.cursor()
        curs.execute("SELECT STAindex,ArrPtr FROM ArrowDirectory WHERE Long='"+arrow+"' OR Short='"+arrow+"' ORDER BY ArrPtr")
        pg_rows = curs.fetchall()
        conn.commit()
        a = pg_rows[0]
        sttype = SST.STIndexToSTType(a[0])
        arrowptr = a[1]
        return arrowptr,sttype

    #

    def STIndexToSTType(sti):        
        return sti - 3
    #    

    def Vertex(conn,name,chapter):
        es = SST.SQLEscape(name)
        ec = SST.SQLEscape(chapter)
        channel = SST.NChannel(name)
        l = len(name)
        
        curs = conn.cursor()
        args =f"{l},{channel},'{es}','{ec}'"
        curs.execute("SELECT IdempAppendNode("+args+")")

        pg_rows = curs.fetchall()
        conn.commit()
        return pg_rows[0]

    #
    
    def Edge(conn,n1,arrowname,n2,context,weight):
        print("MAKE EDGE",n1,arrowname,n2,context,weight)
        arr,sttype = SST.GetDBArrowsWithArrowName(conn,arrowname)
        print("ret",a,s)

#	var link Link
#	link.Arr = arrowptr
#	link.Dst = to.NPtr
#	link.Wgt = weight
#	link.Ctx = RegisterContext(nil,context)


#	AppendDBLinkToNode(ctx,frptr,link,sttype)
#	var invlink Link
#	invlink.Arr = INVERSE_ARROWS[link.Arr]
#	invlink.Wgt = link.Wgt
#	invlink.Dst = frptr
#	AppendDBLinkToNode(ctx,toptr,invlink,-sttype)

#	linkval := fmt.Sprintf("(%d, %f, %d, (%d,%d)::NodePtr)",lnk.Arr,lnk.Wgt,lnk.Ctx,lnk.Dst.Class,lnk.Dst.CPtr)

#	literal := fmt.Sprintf("%s::Link",linkval)

#	link_table := STTypeDBChannel(sttype)

#	qstr := fmt.Sprintf("UPDATE NODE SET %s=array_append(%s,%s) WHERE (NPtr).CPtr = '%d' AND (NPtr).Chan = '%d' AND (%s IS NULL OR NOT %s = ANY(%s))",
#		link_table,
#		link_table,
#		literal,
#		n1ptr.CPtr,
#		n1ptr.Class,
#		link_table,
#		literal,
#		link_table)


#######################################################

class Link:
    
    def __init__(self,arrow,weight,context,nptr):
        self.Arr = arrow
        self.Wgt = weight
        self.Ctx = context
        self.Dst = nptr

#######################################################
# Main
#######################################################

ok,ctx = SST.Open("sstoryline","sst_1234","sstoryline","localhost")

if ok:
    v1 = SST.Vertex(ctx,"first node","examples chapter")
    v2 = SST.Vertex(ctx,"second node","examples chapter")

    print("v1",v1)
    print("v2",v2)
    context = ["freddy","physics"]

    SST.Edge(ctx,v1,"then",v2,context,1)


# Access class and instance variables

n1 = SST.GetDBNodeByNodePtr(ctx,"(3,4)")
n2 = SST.GetDBNodeByNodePtr(ctx,"(3,5)")
print("--",n1)

#

friends = ['john', 'pat', 'gary', 'michael']

for i, name in enumerate(friends):
    
    print (f"iteration {i} is {name}")  # the f fills in the vars
    print ("name:",i,name)

SST.Close(ctx)
