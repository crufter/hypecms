<script src="/shared/nicEdit/nicEdit.js"></script>
<script>
$(function() {

$(".html-editor").each(function(index, elem) {
	var id = $(elem).attr("id")
	//new nicEditor({fullPanel : true, iconsPath : '/shared/nicEdit/nicEditorIcons.gif'}).panelInstance(id)
})

})
</script>
{{if .is_draft}}
	{{if .draft.parent_draft}}
		<a href="/admin/content/edit/{{.type}}_draft/{{.draft.parent_draft}}">Parent draft.</a><br />
		<br />
	{{end}}
	{{if .content_parent}}
		{{if .up_to_date}}
			This draft is up to date.<br />
		{{else}}
			This draft is <b>NOT</b> up to date.<br />
		{{end}}
		<br />
	{{end}}
{{end}}
{{if .is_content}}
	{{if .latest_draft}}
	<a href="/admin/content/edit/{{.type}}_draft/{{.latest_draft._id}}">A fresher draft is available.</a><br />
	<br />
	{{end}}
{{end}}
<form action="/b/content/{{.op}}" method="post">
{{$content := .content}}
{{range .fields}}
	{{.key}}<br />
	{{if eq .key "content"}}
		<textarea id="{{.key}}-field" name="content" class="html-editor"></textarea>
	{{else}}
		<input name="{{.key}}" value="{{.value}}" /><br />
	{{end}}
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
					<a class="delete" href="/b/content/pull_tags?type={{$content.type}}&id={{$content._id}}&tag_id={{._id}}">-</a> {{.name}} ({{.count}})<br /> 
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
<input type="hidden" name="draft_id" value="{{if .is_draft}}{{.draft._id}}{{end}}" />
<input type="hidden" name="id" value="{{if .is_draft}}{{.draft.draft_of}}{{else}}{{$content._id}}{{end}}" />
<input type="submit" name="draft" value="Save as draft"><br />
<br />
<input type="submit" {{if .is_draft}}name="create-content-from-draft"{{end}} value="Save as content">
</form>