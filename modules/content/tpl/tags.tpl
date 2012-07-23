{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>Tags:</h4>
{{range .latest}}
	<a class="delete" href="/b/content/delete_tag?tag_id={{._id}}">-</a> {{.name}} ({{.count}})<br />
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}