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
        ARROW_DIRECTORY.append(arrow)

    curs = conn.cursor()
    curs.execute("SELECT Plus,Minus FROM ArrowInverses ORDER BY Plus")
    pg_rows = curs.fetchall()
    conn.commit()
    for pg_row in pg_rows:
        plus = pg_row[0]
        minus = pg_row[1]
        INVERSE_ARROWS.append(minus)

    return True,conn

#
    
def Close(conn):
    conn.close()

#

def SQLEscape(s):
    return s.replace("'","\\'")

#

def FormatSQLArray(context):
    array = ""
    for i,c in enumerate(context):
        s = SQLEscape(c)
        array += f"\"{s}\""
        if i < len(context)-1:
            array += ","
            
        strarray = "{" + array + "}"
    return strarray
#

def ParseSQLLinkArray(arraystr):
    strarray = arraystr.split(";")
    array = []
    for i,c in enumerate(strarray):
        if len(c) > 0:
            c = c.replace("\"","")
            qtuple = c[1:len(c)-1]
            ftup = qtuple.split(",")
            link = (ftup[0],ftup[1],ftup[2],ftup[3]+","+ftup[4])
            array.append(link)
    return array

#
    
def ParseSQLPathArray(arraystr):
    strarray = arraystr.split("\n")
    patharray = []
    for i,c in enumerate(strarray):
        path = ParseSQLLinkArray(c)
        if len(path) > 0:
            patharray.append(path)
    return patharray

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
    sttype = STIndexToSTType(a[0])
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
    if pg_rows:
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
    ctxstr = NormalizeContext(context)
    if len(ctxstr) == 0:
        return
    str,ctxptr = GetDBContextByName(conn,ctxstr)
    if ctxptr == -1 or str != ctxstr:
        ctxptr = UploadContextToDB(ctx,ctxstr,-1)
    return ctxptr

#    

def STIndexToSTType(sti):        
    return sti - 3

#    

def Vertex(conn,name,chapter):
    es = SQLEscape(name)
    ec = SQLEscape(chapter)
    channel = NChannel(name)
    l = len(name)        
    curs = conn.cursor()
    args =f"{l},{channel},'{es}','{ec}'"
    curs.execute("SELECT IdempAppendNode("+args+")")
    pg_rows = curs.fetchall()
    conn.commit()
    return pg_rows[0][0]

#
    
def Edge(conn,n1,arrowname,n2,context,weight):
    arr,sttype = GetDBArrowsWithArrowName(conn,arrowname)
    ctxptr = TryContext(conn,context)
    link = (arr,weight,ctxptr,n2)
    AppendDBLinkToNode(conn,n1,link,sttype)

#
    
def AppendDBLinkToNode(conn,frptr,link,sttype):
    arr = link[0]
    weight = link[1]
    ctxptr = link[2]
    dst = link[3]
    txtlink = f"({arr},{weight},{ctxptr},{dst}::NodePtr)::Link"
    Ix = STTypeDBChannel(sttype)

    cmd = f"UPDATE NODE SET {Ix}=array_append({Ix},{txtlink}) WHERE NPtr='{frptr}'::NodePtr AND ({Ix} IS NULL OR NOT {txtlink} = ANY({Ix}))"
    curs = conn.cursor()
    curs.execute(cmd)
    conn.commit()

    invarr = INVERSE_ARROWS[arr]
    invlink = f"({invarr},{weight},{ctxptr},{frptr}::NodePtr)::Link"
    invIx = STTypeDBChannel(-sttype)
    icmd = f"UPDATE NODE SET {invIx}=array_append( {invIx} , {invlink} ) WHERE NPtr='{dst}'::NodePtr AND ( {invIx} IS NULL OR NOT {invlink} = ANY({invIx}) )"
    curs = conn.cursor()
    curs.execute(icmd)
    conn.commit()

#
        
def GetFwdPathsAsLinks(conn,nptr,sttype,depth,maxlimit):

    cmd = f"SELECT FwdPathsAsLinks from FwdPathsAsLinks('{nptr}',{sttype},{depth},{maxlimit})"
    curs = conn.cursor()
    curs.execute(cmd)
    pg_rows = curs.fetchall()
    conn.commit()
    for pg_row in pg_rows:
        if len(pg_row[0]) > 0:
            links = ParseSQLPathArray(pg_row[0])
    return links,len(links)

#

def GetEntireNCConePathsAsLinks(conn,orientation,nptrarr,depth,chapter,contex,limit):
    startset = FormatSQLArray(nptrarr)
    context = FormatSQLArray(contex)
    cmd = f"SELECT AllNCPathsAsLinks('{startset}','%{chapter}%',false,'{context}','{orientation}',{depth},{limit})"
    curs = conn.cursor()
    curs.execute(cmd)
    pg_rows = curs.fetchall()
    conn.commit()
    for pg_row in pg_rows:
        if len(pg_row[0]) > 1:
            links = ParseSQLPathArray(pg_row[0])
    return links,len(links)

