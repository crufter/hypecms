{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>All enries: </h4>
{{require content/search-form.t}}
{{range .latest}}
	{{require content/listing.t}}
{{end}}

<br />
{{$navi := .paging}}
{{require admin/navi.t}}

{{require content/footer.t}}
{{require admin/footer.t}}