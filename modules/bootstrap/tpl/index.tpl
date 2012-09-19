{{require admin/header.t}}
<div id="middle-content">
	{{if .not_admin}}
		<h3>This is not an admin instance.</h3>
		This means people will not be able to register sites at you.<br />
		<br />
	{{end}}
	<a href="/b/bootstrap/start-all">Start all sites.</a><br />
	<br />
	<form>
		<input type="text" name="search">
		<input type="button" value="Search">
	</form>
	{{if eq .match .all}}
		{{.all}} sites.
	{{else}}
		{{.match}}/{{.all}} sites.
	{{end}}
	<br />
	<br />
	{{range .sitenames}}
		<a class="delete" href="/b/bootstrap/delete-site?sitename={{.}}">-</a> {{.}}<br />
	{{end}}
</div>
{{require admin/footer.t}}