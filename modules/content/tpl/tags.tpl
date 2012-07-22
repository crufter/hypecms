{{require admin/header.t}}
{{require content/sidebar.t}}

{{.type}} entries:<br /><br />
{{range .latest}}
	{{.name}} ({{.count}}) <a href="/b/content/delete_tag?tag_id={{._id}}">x</a><br />
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}