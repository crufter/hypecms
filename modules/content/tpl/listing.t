<div class="list-item">
	<a class="delete" href="/b/content/delete?id={{._id}}&type={{.type}}" />-</a>
	<a href="/admin/content/edit?type={{.type}}&id={{._id}}">{{if .title}}{{.title}}{{else}}{{.name}}{{end}}</a>
	<!-- Has up to date draft. -->
	{{if .latest_draft}}
		<span class="info"><a href="/admin/content/edit/{{.latest_draft.type}}/{{.latest_draft._id}}">Has draft</a></span>
	{{end}}
</div>