{{require admin/header.t}}
{{require content/sidebar.t}}

You gotta see a list of {{.type}} content here.<br />
{{range .latest}}
	{{require content/listing.t}}
{{end}}

{{require content/footer.t}}
{{require admin/footer.t}}