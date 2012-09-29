{{require header.t}}

<h1>New:</h1>
<form action="{{action "insert"}}">
{{range .scheme}}
		{{.key}}<br />
		<input name="{{key .key}}"/><br />
		<br />
{{end}}
<input type="submit" />
</form>

{{require footer.t}}