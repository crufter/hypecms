<a class="delete" href="/b/content/delete?id={{._id}}&type={{.type}}" />-</a>
{{if .title}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.title}}</a>
{{else}}
	<a href="/admin/content/edit/{{.type}}/{{._id}}">{{.name}}</a>
{{end}}
<!-- Has up to date draft. -->
{{if .latest_draft}}
	<span class="info"><a href="/admin/content/edit/{{.latest_draft.type}}/{{.latest_draft._id}}">Has draft</a></span>
{{end}}
<br />