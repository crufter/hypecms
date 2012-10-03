{{require header.t}}

<h1>New:</h1>
<!--
	{{$f := form "insert"}}
	{{$f.ActionPath}}<br />
	{{$f.FilterFields}}<br />
	{{$f.KeyPrefix}}<br />
	<br />
-->
<form action="{{$f.ActionPath}}" method="POST">
	{{$f.HiddenString}}
	{{range .scheme}}
			{{.key}}<br />
			<input name="{{$f.KeyPrefix}}{{.key}}"/><br />
			<br />
	{{end}}
	<input type="submit" />
</form>

{{require footer.t}}