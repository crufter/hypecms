<form action="/b/content/{{.op}}" method="post">
{{$content := .content}}
{{range .fields}}
	{{.key}}<br />
	<input name="{{.key}}" value="{{.value}}" /><br />
	<br />
	{{if .tags}}
		<script src="/tpl/content/tag_finder.js"></script>
		<style>
		#autocomplete{
			display: none;
			background: #f8f8f8;
			position: absolute;
			border-left: 1px solid #ccc;
			border-right: 1px solid #ccc;
			border-bottom: 1px solid #ccc;
			box-shadow: 0px 0px 5px #888;
		}
		.tag-option{
			padding: 0.5em 1em;
			cursor: pointer;
		}
		.tag-option:hover{
			background: #e8e8e8;
		}
		.selected{
			background: #cacaca;
		}
		</style>
		{{if $content._tags}}
			{{range $content._tags}}
				{{if .}}
					<a class="delete" href="/b/content/pull_tags?content_id={{$content._id}}&tag_id={{._id}}">-</a> {{.name}} ({{.count}})<br /> 
				{{end}}
			{{end}}
			<br />
		{{else}}
			No tags yet.<br />
			<br />
		{{end}}
	{{end}}
{{end}}
<input type="hidden" name="type" value="{{.type}}" />
{{if .is_draft}}
	<input type="hidden" name="prev_draft" value="{{if .draft_parent}}{{$content._id}}{{end}}" /> <!-- For first draft-from-content you do not need to send it. -->
	<input type="hidden" name="id" value="{{if .content_parent}}{{$content._id}}{{end}}" />
{{else}}
	<input type="hidden" name="id" value="{{.content._id}}" />
{{end}}
<input type="submit" name="draft" value="Save as draft"><br />
<br />
<input type="submit">