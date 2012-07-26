{{require admin/outer-header.t}}

<h3>Register as administrator</h3>
Since the website has no admin yet, please provide your admin password here:<br />
<br />
<form action="/admin/b/regfirstadmin">
	Password:<br />
	<input name="password" type="password"><br />
	<br />
	Password again:<br />
	<input name="password_again" type="password"><br />
	<br />
	<input type="submit">
</form>

{{require admin/outer-footer.t}}