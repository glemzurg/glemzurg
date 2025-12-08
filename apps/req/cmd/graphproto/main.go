package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/goccy/go-graphviz"
)

var _dot = `


digraph UML {
    graph [splines=ortho, nodesep=0.5, ranksep=1];
    node [shape=plaintext, fontname="Sans-Serif", fontsize=11];
    edge [arrowhead=none, fontname="Sans-Serif", fontsize=10];// Force horizontal layout for the association with assoc class
{rank=same; ClassA; dummy; ClassC;}

// Example classes
ClassA [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><I>&laquo;Interface&raquo;</I><BR/><B>ClassA</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- attribute1: String</TD></TR>
        <TR><TD ALIGN="LEFT">- attribute2: int</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ operation1(): void</TD></TR>
        <TR><TD ALIGN="LEFT">+ operation2(param: bool): String</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

ClassB [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><B>ClassB</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- attrX: float</TD></TR>
        <TR><TD ALIGN="LEFT">- attrY: Date</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ methodX(): int</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

ClassC [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><I>&laquo;Abstract&raquo;</I><BR/><B>ClassC</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- field1: bool</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ abstractMethod(): void</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

AssocClass [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><B>AssocClass</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- assocAttr: String</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ assocOp(): void</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

// Added inheritance examples
ClassD [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><B>ClassD</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- specificAttr: double</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ concreteMethod(): bool</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

ClassE [label=<<TABLE BORDER="1" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4" WIDTH="200">
    <TR><TD ALIGN="CENTER"><B>ClassE</B></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">- implAttr: String</TD></TR>
    </TABLE></TD></TR>
    <TR><TD ALIGN="LEFT"><TABLE BORDER="0" CELLBORDER="0" CELLSPACING="0" WIDTH="100%">
        <TR><TD ALIGN="LEFT">+ implOperation(): void</TD></TR>
    </TABLE></TD></TR>
</TABLE>>];

// Inheritance edges
ClassD -> ClassC [arrowhead=empty];
ClassE -> ClassA [arrowhead=empty, style=dashed];

// Simple association example between ClassA and ClassB - arrow at ClassB
ClassA -> ClassB [taillabel="1", headlabel="0..1", label="relatesTo", arrowhead=normal, labeldistance=2.0];

// Association between ClassB and ClassC - arrow at ClassB (reversed direction)
ClassC -> ClassB [taillabel="1..*", headlabel="*", label="contains", arrowhead=normal];

// Association with association class - arrow at ClassC
dummy [shape=point, width=0.01, height=0.01, label=""];
ClassA -> dummy [taillabel="1..*", minlen=1];
dummy -> ClassC [headlabel="*", label="owns", minlen=1, arrowhead=normal];
AssocClass -> dummy [style=dashed];}

`

// var _dot = `

// digraph UML_Class_diagram {

// 	graph [
// 		label="UML Class diagram demo"
// 		labelloc="t"
// 		fontname="Helvetica,Arial,sans-serif"
// 	]
// 	node [
// 		fontname="Helvetica,Arial,sans-serif"
// 		shape=record
// 		style=filled
// 		fillcolor=gray95
// 	]
// 	edge [fontname="Helvetica,Arial,sans-serif"]
// 	edge [arrowhead=vee style=dashed]
// 	Client -> Interface1 [label=dependency]
// 	Client -> Interface2

// 	edge [dir=back arrowtail=empty style=""]
// 	Interface1 -> Class1 [xlabel=inheritance]
// 	Interface2 -> Class1 [dir=none]
// 	Interface2 [label="" xlabel="Simple\ninterface" shape=circle]

// 	Interface1[label = <{<b>«interface» I/O</b> | + property<br align="left"/>...<br align="left"/>|+ method<br align="left"/>...<br align="left"/>}>]
// 	Class1[label = <{<b>I/O class</b> | + property<br align="left"/>...<br align="left"/>|+ method<br align="left"/>...<br align="left"/>}>]
// 	edge [dir=back arrowtail=empty style=dashed]
// 	Class1 -> System_1 [label=implementation]
// 	System_1 [
// 		shape=plain
// 		label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
// 			<tr> <td> <b>System</b> </td> </tr>
// 			<tr> <td>
// 				<table border="0" cellborder="0" cellspacing="0" >
// 					<tr> <td align="left" >+ property</td> </tr>
// 					<tr> <td port="ss1" align="left" >- Subsystem 1</td> </tr>
// 					<tr> <td port="ss2" align="left" >- Subsystem 2</td> </tr>
// 					<tr> <td port="ss3" align="left" >- Subsystem 3</td> </tr>
// 					<tr> <td align="left">...</td> </tr>
// 				</table>
// 			</td> </tr>
// 			<tr> <td align="left">+ method<br/>...<br align="left"/></td> </tr>
// 		</table>>
// 	]

// 	edge [dir=back arrowtail=diamond]
// 	System_1:ss1 -> Subsystem_1 [xlabel="composition"]

// 	Subsystem_1 [
// 		shape=plain
// 		label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
// 			<tr> <td> <b>Subsystem 1</b> </td> </tr>
// 			<tr> <td>
// 				<table border="0" cellborder="0" cellspacing="0" >
// 					<tr> <td align="left">+ property</td> </tr>
// 					<tr> <td align="left" port="r1">- resource</td> </tr>
// 					<tr> <td align="left">...</td> </tr>
// 				</table>
// 				</td> </tr>
// 			<tr> <td align="left">
// 				+ method<br/>
// 				...<br align="left"/>
// 			</td> </tr>
// 		</table>>
// 	]
// 	Subsystem_2 [
// 		shape=plain
// 		label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
// 			<tr> <td> <b>Subsystem 2</b> </td> </tr>
// 			<tr> <td>
// 				<table align="left" border="0" cellborder="0" cellspacing="0" >
// 					<tr> <td align="left">+ property</td> </tr>
// 					<tr> <td align="left" port="r1">- resource</td> </tr>
// 					<tr> <td align="left">...</td> </tr>
// 				</table>
// 				</td> </tr>
// 			<tr> <td align="left">
// 				+ method<br/>
// 				...<br align="left"/>
// 			</td> </tr>
// 		</table>>
// 	]
// 	Subsystem_3 [
// 		shape=plain
// 		label=<<table border="0" cellborder="1" cellspacing="0" cellpadding="4">
// 			<tr> <td> <b>Subsystem 3</b> </td> </tr>
// 			<tr> <td>
// 				<table border="0" cellborder="0" cellspacing="0" >
// 					<tr> <td align="left">+ property</td> </tr>
// 					<tr> <td align="left" port="r1">- resource</td> </tr>
// 					<tr> <td align="left">...</td> </tr>
// 				</table>
// 				</td> </tr>
// 			<tr> <td align="left">
// 				+ method<br/>
// 				...<br align="left"/>
// 			</td> </tr>
// 		</table>>
// 	]
// 	System_1:ss2 -> Subsystem_2;
// 	System_1:ss3 -> Subsystem_3;

// 	edge [xdir=back arrowtail=odiamond]
// 	Subsystem_1:r1 -> "Shared resource" [label=aggregation]
// 	Subsystem_2:r1 -> "Shared resource"
// 	Subsystem_3:r1 -> "Shared resource"
// 	"Shared resource" [
// 		label = <{
// 			<b>Shared resource</b>
// 			|
// 				+ property<br align="left"/>
// 				...<br align="left"/>
// 			|
// 				+ method<br align="left"/>
// 				...<br align="left"/>
// 			}>
// 	]
// }

// `

func main() {

	// Example call: go run main.go > graph.svg

	ctx := context.Background()
	g, err := graphviz.New(ctx)
	if err != nil {
		panic(err)
	}

	graph, err := graphviz.ParseBytes([]byte(_dot))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := graph.Close(); err != nil {
			panic(err)
		}
		g.Close()
	}()

	var buf bytes.Buffer
	if err := g.Render(ctx, graph, graphviz.SVG, &buf); err != nil {
		log.Fatal(err)
	}
	fmt.Println(buf.String())
}
