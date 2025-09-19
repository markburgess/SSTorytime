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

    # We do not allow Vertex/Edge users to define arrows.
    ARROW_DIRECTORY = []
    INVERSE_ARROWS = []

    def __init__(self):
        return

    def Open(dbuser,dbpwd,dbname,dbhost):

        try:
            conn = psycopg2.connect(database=dbname,user=dbuser,password=dbpwd,host=dbhost,port=5432)
        except:
            print("failed")
            return False,conn

        # Download arrows and contexts

        curs = conn.cursor()
        curs.execute("SELECT STAindex,Long,Short,ArrPtr FROM ArrowDirectory ORDER BY ArrPtr")
        pg_rows = curs.fetchall()
        conn.commit()
        for pg_row in pg_rows:
            arrow = ( pg_row[0],pg_row[1],pg_row[2],pg_row[3] )
            SST.ARROW_DIRECTORY.append(arrow)

        curs = conn.cursor()
        curs.execute("SELECT Plus,Minus FROM ArrowInverses ORDER BY Plus")
        pg_rows = curs.fetchall()
        conn.commit()
        for pg_row in pg_rows:
            arrow = ( pg_row[0],pg_row[1] )
            SST.INVERSE_ARROWS.append(arrow)

        return True,conn

    #
    
    def Close(conn):
        conn.close()

    #

    def SQLEscape(s):
        return s.replace("'","\\'")

    #

    def STTypeDBChannel(sttype):
        match sttype:
            case -3:
                return "Im3"
            case -2:
                return "Im2"
            case -1:
                return "Im1"
            case 0:
                return "In0"
            case 1:
                return "Il1"
            case 2:
                return "Ic2"
            case 3:
                return "Ie3"
            
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

    def NormalizeContext(array):
        adict = {}
        ordered = ""
        for part in array:
            adict[part] = 1
        akeys = list(adict.keys())
        akeys.sort()
        for index,part in enumerate(akeys):
            ordered += part
            if index < len(part)-1:
                ordered = ordered + ","
        return ordered

    #
    
    def GetDBContextByName(conn,ctxstr):
        curs = conn.cursor()
        cmd = f"SELECT DISTINCT Context,CtxPtr FROM ContextDirectory WHERE Context='{ctxstr}'"
        curs.execute(cmd)
        pg_rows = curs.fetchall()
        conn.commit()
        if 'pg_rows' in locals():
            return pg_rows[0][0],pg_rows[0][1]
        else:
            return "",-1
    #

    def UploadContextToDB(conn,ctxstr,cptr):
        curs = conn.cursor()
        if len(ctxstr) == 0:
            return
        cmd = f"SELECT IdempInsertContext('{ctxstr}',{cptr})"
        curs.execute(cmd)
        pg_rows = curs.fetchall()
        conn.commit()        
        return
    
    #
    
    def TryContext(conn,context):
        ctxstr = SST.NormalizeContext(context)
        if len(ctxstr) == 0:
            return
        str,ctxptr = SST.GetDBContextByName(conn,ctxstr)
        if ctxptr == -1 or str != ctxstr:
            ctxptr = SST.UploadContextToDB(ctx,ctxstr,-1)
        return ctxptr
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
        return pg_rows[0][0]

    #
    
    def Edge(conn,n1,arrowname,n2,context,weight):
        #print("MAKE EDGE",n1,arrowname,n2,context,weight)
        arr,sttype = SST.GetDBArrowsWithArrowName(conn,arrowname)
        print("ret",arr,sttype)
        ctxptr = SST.TryContext(conn,context)
        link = f"({arr},{weight},{ctxptr},{n2}::NodePtr)"
        SST.AppendDBLinkToNode(ctx,n1,link,sttype)

    #
    
    def AppendDBLinkToNode(ctx,frptr,link,sttype):
        Ix = SST.STTypeDBChannel(sttype)
        cmd = f"UPDATE NODE SET {Ix}=array_append({Ix},{link}) WHERE NPtr='{frptr}' AND (Ix IS NULL OR NOT {link} = ANY(Ix))",        

        invarr = SST.INVERSE_ARROWS[arr]
        invlink = f"({invarr},{weight},{ctxptr},{link[3]}::NodePtr)"
        invIx = SST.STTypeDBChannel(-sttype)
        icmd = f"UPDATE NODE SET {invIx}=array_append({invIx},{invlink}) WHERE NPtr='{link[3]}' AND (invIx IS NULL OR NOT {invlink} = ANY(invIx))",        

        print("FWD",cmd)
        print("BWD",icmd)

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

    print(f"v1 {v1}")
    print("v2",v2)
    context = ['dunnum', 'cotton', 'pickin','lumberjack']

    SST.Edge(ctx,v1,"then",v2,context,1)


# Access class and instance variables

n1 = SST.GetDBNodeByNodePtr(ctx,"(3,4)")
n2 = SST.GetDBNodeByNodePtr(ctx,"(3,5)")
print("--",n1)

#

friends = ['john', 'pat', 'gary', 'michael']

print("......",SST.NormalizeContext(friends))

for i, name in enumerate(friends):
    
    print (f"iteration {i} is {name}")  # the f fills in the vars
    print ("name:",i,name)

SST.Close(ctx)
