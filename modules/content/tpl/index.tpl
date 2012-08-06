{{require admin/header.t}}
{{require content/sidebar.t}}

<h4>All enries: </h4>
{{require content/search-form.t}}
{{range .latest}}
	{{require content/listing.t}}
{{end}}

<br />
<div class="navi">
	{{range .paging}}
		<a href="{{.Url}}">{{.Page}}</a> 
	{{end}}
</div>

{{require content/footer.t}}
{{require admin/footer.t}}