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

{{require content/footer.t}}
{{require admin/footer.t}}