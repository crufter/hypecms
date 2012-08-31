{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>
{{if .is_content}}
	{{if .content._id}}
		Edit
	{{else}}
		Insert
	{{end}} {{.type}} content
{{end}}
{{if .is_draft}}
	Edit {{.type}} draft
{{end}}
{{if .is_version}}
	Not implemented yet.
{{end}}
</h4>
{{require content/edit-form.t}}

<br />

<script src="/shared/arborjs/lib/arbor.js"></script>
<script src="/shared/arborjs/lib/arbor-tween.js"></script>
<script src="/tpl/content/main.js"></script>

<div>
	<canvas id="viewport" width="700" height="500" style="border: 1px solid #ccc;"></canvas>
</div>

<br />
Click on a node.
<br /><br />
<div id="info">

</div>

<script>
$(function(){
  function newNode(id, type) {
	this.id = id
	this.type = type
  }
  
   var last_selected
  newNode.prototype.Select = function() {
	if (last_selected != undefined) {
		last_selected.label = last_selected.label.slice(2, -2)
	}
	this.label = "[ " + this.label + " ]"
	last_selected = this
  }
  
  newNode.prototype.PrintData = function() {
	var type
	switch (this.type) {
	case 0:
		type = "{{.type}}_draft"
	break
	case 1:
		type = "{{.type}}_version"
	break
	case 2:
		type = {{.type}}
	break
	}
	var url = "/"
	$("#info").html('<a href="/admin/content/edit/'+type+'/'+this.id+'">Link</a>')
  }
  
    var sys = arbor.ParticleSystem(1000, 600, 0.5) // create the system with sensible repulsion/stiffness/friction
    sys.parameters({gravity:true}) // use center-gravity to make the graph settle nicely (ymmv)
    sys.renderer = Renderer("#viewport") // our newly created renderer will have its .init() method called shortly by sys...

    // add some nodes to the graph and watch it go...
	var data = {{.timeline}}
	var nodes = {}
	var version_counter = 1
	var draft_counter = 1
	for (var i in data) {
		var c = data[i]
		var type	// 0 draft, 1 version, 2 head
		if (c["data"] != undefined) {
			type = 0
		} else if (c["version_date"] != undefined) {
			type = 1
		} else {
			type = 2
		}
		var id = c["_id"]
		var node = new newNode(id, type)
		switch (type) {
			case 0:
				node["label"] = "d" + draft_counter++
				node["color"] = "grey"
			break
			case 1:
				node["label"] = "v" + version_counter++
				node["color"] = "blue"
			break
			case 2:
				node["label"] = "h"
				node["color"] = "red"
			break
		}
		nodes[id] = node
	}
	var edges = {}
	for (var i in data) {
		var c = data[i]
		var destination
		if (c["-parent"] != undefined) {
			destination = "-parent"
		} else if (c["pointing_to"] != undefined) {
			destination = "pointing_to"
		} else {
			continue
		}
		if (edges[c[destination]] == undefined) {
			edges[c[destination]] = {}
		}
		var edge = {}
		var id = c["_id"]
		edge[c[destination]] = {}
		edges[id] = edge
	}
	console.log(nodes, edges)
	sys.graft({"nodes": nodes, "edges": edges})
	
    //sys.addEdge('a','b')
    //sys.addEdge('a','c')
    //sys.addEdge('a','d')
    //sys.addEdge('a','e')
    //sys.addNode('f', {alone:true, mass:.25})

    // or, equivalently:
    //
    // sys.graft({
    //   nodes:{
    //     f:{alone:true, mass:.25}
    //   }, 
    //   edges:{
    //     a:{ b:{},
    //         c:{},
    //         d:{},
    //         e:{}
    //     }
    //   }
    // })
})
</script>

{{require content/footer.t}}
{{require admin/footer.t}}