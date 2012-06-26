{{require admin/header.t}}
{{require content/sidebar.t}}

You gotta see a list of {{.type}} content here.
{{range .latest}}
	{{if .title}}
		{{.title}}
	{{else}}
		{{.name}}
	{{end}}<br />
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}