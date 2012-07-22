<form action="/b/content/{{.op}}" method="post">
{{range .fields}}
	{{.key}}<br />
	<input name="{{.key}}" value="{{.value}}" /><br />
	<br />
{{end}}
{{if .tags_on}}
	<script src="/tpl/content/tag_finder.js"></script>
	<style>
	#autocomplete{
		padding: 2px 5px;
		border-left: 1px solid #ccc;
		border-right: 1px solid #ccc;
		border-bottom: 1px solid #ccc;
		box-shadow: 0px 0px 5px #888;
	}
	.tag-option{
		padding: 2px;
		cursor: pointer;
	}

	.tag-option:hover{
		background: #e8e8e8;
	}
	.selected{
		background: #cacaca;
	}
	</style>
	{{$content_id := .content._id}}
	{{if .content._tags}}
		{{range .content._tags}}
			{{.name}} ({{.count}}) <a href="/b/content/pull_tags?content_id={{$content_id}}&tag_id={{._id}}">x</a> <br /> 
		{{end}}
		<br />
	{{else}}
		No tags yet.<br />
		<br />
	{{end}}
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
<input type="hidden" name="id" value="{{.content._id}}" />
<input type="submit">